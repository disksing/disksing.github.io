---
title: "PD leader 切换耗时分析"
date: "2020-02-19"
tags: ["TiDB"]
---

> 本文首发在[AskTUG.com](https://asktug.com/t/topic/2767)。

我们知道，TiDB 集群中的多个 PD 提供的服务的方式是选出一个 PD 作为 leader 来提供服务的，当 leader 出现故障或者网络隔离后，其余的节点会自动通过 raft 选出新的 leader 继续服务。

从旧 leader 故障，到新 leader 选出并开始提供服务，这个过程服务是不可用的（比如 TSO 服务不可用，导致事务被 block），所以有必要分析这个过程的耗时并尽量使其缩短。

值得注意的是，PD 的配置有个相关的参数叫 lease，也就是 leader 的租约期限，默认值是 3s。那么，是否 leader 切换的耗时就是 3 秒呢？

实际我们观察到的往往比这个配置值要长不少，可能要 10 多秒甚至好几十秒，下面我们就来分析一下，时间都去哪儿了。

## 流程分析

**1. etcd Leader 选举**

PD 的 leader 选举并非自己实现 raft，而是直接内嵌了 [embedded etcd](https://github.com/etcd-io/etcd/tree/master/embed)，然后基于 etcd 提供的 lease 机制来实现 leader 选举。

内嵌 etcd 的好处是部署上比较简单，但是也带来了负面效果：如果 etcd 的 leader 在故障的节点上，PD 需要先等 etcd 选举出 leader 并恢复服务。

PD 配置中的 `election-interval` 用于控制 etcd 的选举超时时间，默认配置也是 3s。通常情况下，这一步耗时约为 3s，如果选举时出现分票的情况，可能还会稍长一些。

**2. PD leader lease 过期**

PD 竞选 leader 的机制是抢占式地往 etcd 的特定的 leader key 写入自己的 member id，并向 etcd 注册 lease。只要 leader 在线，会不断地更新 lease，也就能维持自己的 leader 角色。一旦发生故障或者隔离，etcd 的 LeaseManager 会在 lease 过期后删除这个 key，其他 PD 节点 watch 到 leader key 被删除之后会尝试把自己注册成新的 leader。

etcd 所管理的 lease 是在内存中倒计时的，不会实时地把剩余时间写入 raft 状态机，而是只存放了 TTL 信息。这带来一个问题，当 etcd leader 故障，新的节点成为 leader 后，无法得知之前的 lease 消耗了多少了，只能从头开始倒计时。而且 etcd 要恢复 leasor 时还会[多加一个 Election timeout](https://github.com/etcd-io/etcd/blob/d6a3c995cf86b479cb5a44b48d000feb33e3d8f8/etcdserver/server.go#L1976-L1980)，（这里不太理解，可能是出于某种安全性的考虑？）

```
// promote lessor when the local member is leader and finished
// applying all entries from the last term.
if s.isLeader() {
    s.lessor.Promote(s.Cfg.electionTimeout())
}
```

这样一来，TTL+electionTimeout 至少就有 6 秒了。实际测试这一步大约耗时在 6-8s。

**3. PD 竞选 leader**

当 PD watch 到 leader key 被 lease manager 删掉之后，则进入竞选状态，尝试将 leader key 设为自己的 member ID 并设置 lease。这一步通常很快就能完成。

**4. TSO 时钟同步**

PD 竞选成功后不能立即开始服务，需要确保分配的 ts 不能小于之前 leader 分配的 ts。首先 PD 会从 etcd 读出上一个 leader 可能分配过的最大 ts，接着检查本地时钟确保大于之前的 ts（如果不同 PD 之间时钟不同步，会需要 sleep 等待），最后再把当前时间+3s（可通过 `tso-save-interval` 调整）作为新的“可以分配的最大 ts” 持久化 etcd。

这一步的时间主要取决于时钟不同步的程度，如果正常开启 ntp 的话很快就能完成。

**5. 元信息加载**

除了 TSO，PD 还要为 TiDB 提供 Region 信息查询的功能。因此，PD 在开始服务之前，需要把所有 Regoin 元信息加载进内存。

这一步的时间主要跟 Region 的数量相关，如果集群规模不大对整体耗时的影响比较小，但是对于几百万 Region 的大集群，可能会需要长达几十秒。

### 相关优化

切换 leader 的耗时是很值得优化的，毕竟作为集群的单点，PD 不可用的影响范围是整个集群。这里简单列举一下优化手段（包括规划中的）：

**1. 调小相关的 timeout 参数**

这个是很直接的方法，但是要注意过小的 timeout 可能会导致瞬间的网络抖动产生重新选举，尤其是跨机房的场景。

**2. 开启 PreVote**

这个主要是针对 etcd leader 选举的过程，开启 PreVote 可以使 leader 选举更稳定，降低选举时发生分票的概率。2.0.x 以上版本都是默认开启了 PreVote 的。

**3. TSO 不受系统时间影响**

这个优化其实不只针对不同节点间时间不一致的情况，也能解决手动调整系统时间产生的 TSO 暂停服务的问题。

简单来说，当可分配的 TS 超出当前的系统时间时，不再是直接停止服务，而是以很慢的速度递增 TS 的 physical 部分，直到系统时间追上来。

这个优化包含在 2.1.0 以上版本中。

**4. TSO 提前开始提供服务**

在 leader 切换的过程中，相对于 Region 信息服务，TSO 对可用性的影响更为致命，因为 TiDB 往往缓存了常用的 Region 信息，而且一部 Region 信息也能从 TiKV 更新，而 TSO 是没有任何旁路冗余的。

很自然的想法就是当 TSO 时钟同步完成后，不必等元信息加载完成，TSO 可以提前开始服务，这样可以一定程度上降低 leader 切换对整个集群的影响。

这个优化包含在 3.1.0 以上版本中。

**5. 提前加载 Region 元信息**

在 3.0.0 版本之后，为了支持大规模集群，PD 不再把 Region 存放在 etcd 中了，而是自行存储在文件系统并通过 `RegionSyncer` 组件同步给 Follower。

我们可以借助 `RegionSyncer` 在 Follower 上提前准备好 Region 的内在数据结构，等到 Follower 变成 Leader 之后，直接启用就行了，不需要加载的过程。

这个优化将在 4.0.0 版本提供出来。

**6. 不依赖于 etcd 自己实现 lease 机制**

有了上面那些优化之后，耗时的大头就落在前两个步骤了，也就是 etcd leader 选举和等待 leader lease 过期，它们加起来的耗时就可能达到 10 多秒。因此想进一步优化 leader 切换的耗时，不得不啃掉这块硬骨头了。

目前一个吸引了我们注意的方案是[Paxos lease](https://arxiv.org/pdf/1209.4187.pdf)，不过这个事情要最终落地恐怕也没那么快，只能说值得期待吧。

