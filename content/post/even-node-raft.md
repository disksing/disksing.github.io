---
title: "偶数节点 raft"
date: "2020-01-31"
tags: ["分布式系统", "共识算法", "raft"]
description: "为奇葩方案平个反"
---

对 raft 有所了解的同学都知道，raft 一般会使用奇数个节点，比如 3，5，7 等等。这是因为 raft 是 一种基于多节点投票选举机制的共识算法，通俗地说，只有超过半数节点在线才能提供服务。这里*超过半数*的意思是***N/2+1***（而不是*N/2*），举例来说，3 节点集群需要 2 个以上节点在线，5 节点集群需要 3 个以上节点在线，等等。对于偶数节点的集群，2 节点集群需要 2 节点同时在线，4 节点集群需要 3 节点在线，以此类推。实际上不只是 raft，所有基于 Quorum 的共识算法大体上都是这么个情况，例如 Paxos，ZooKeeper 什么的，本文仅以 raft 为例讨论。

先考察一下为什么 raft 通常推荐使用奇数节点而不是偶数节点。

共识算法要解决的核心问题是什么呢？是分布式系统中单个节点的不可靠造成的不可用或者数据丢失。raft 保存数据冗余副本来解决这两个问题，当少数节点发生故障时，剩余的节点会自动重新进行 leader 选举（如果需要）并继续提供服务，而且 log replication 流程也保证了剩下的节点（构成 Quorum）总是包含了故障前成功写入的最新数据，因此也不会发生数据丢失。

我们对比一下 3 节点的集群和 4 节点的集群，Quorum 分别是 2 和 3，它们能容忍的故障节点数都是 1。如果深究的话，从概率上来说 4 节点集群发生 2 节点同时故障的可能性要更高一些。于是我们发现，相对于 3 节点集群，4 节点集群消耗更多的硬件资源，却换来了更差的可用性，显然不是个好选择。

但是！！！

上面说了，raft 解决的核心问题有两个，分别是高可用和数据容灾。跟奇数节点相比，偶数节点的方案从可用性上看很不划算，但是数据容灾方面却是有优势的。还是以 4 节点为例，因为 Quorum 是 3，写入数据的时候需要复制到至少 3 个节点才算写入成功，假如此时有 2 个节点同时故障，这种情况下虽然不可用了，但是剩余的两个节点一定包含有最新的数据，因此没有发生数据丢失。这一点很容易被忽视，在常见的奇数节点配置下，保证可用和保证数据不丢所容忍的故障节点数是重合的，但是在偶数节点配置下是不一样的。

根据上面的分析，偶数节点集群的适用场景是“能容忍一定时间的不可用，但不能容忍数据丢失”，应该有不少严肃的金融场景是符合这个描述的，毕竟一段时间不服务也比丢掉数据要强呀。

下面以两数据中心环境为例来对比一下。限制条件是任意一个数据中心故障时（比如发生严重自然灾害），能容忍一定时间的不可用，但不允许发生数据丢失。

如果使用奇数节点集群配置，两个数据中心的节点数一定是不对等的，一旦节点数更多的那个数据中心故障，就可能发生数据丢失了。而如果使用偶数节点配置，两个数据中心的节点数是一样的，任意一个数据中心故障后，另一个数据中心一定包含有最新数据，我们只需要使用工具改写 raft 元信息，让剩余数据中心的所有节点组成新的 raft group 并使得 Quorum 恰好等于剩余节点数，raft 选举机制将会自动选择包含有最新数据的节点当 leader 并恢复服务。

-----

题外话：本来想在 etcd 上实践下这套方案，可惜最后一步 etcd 恢复数据的时候只支持从单一节点恢复，所以无法做到“自动选择包含有最新数据的节点当 leader 并恢复服务”。我[给 etcd 提了个 issue](https://github.com/etcd-io/etcd/issues/11486) 不过貌似并没有成功让他们了解到我想干啥，如果有人看到这里觉得这事情有搞头的话，可以帮忙去 issue 下支持一下。。。
