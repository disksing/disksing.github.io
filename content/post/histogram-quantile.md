---
title: "histogram_quantile 相关的若干问题"
date: "2020-01-30"
tags: ["分布式系统"]
description: "可能是 Prometheus 最难理解的概念"
---

`histogram_quantile` 是 Prometheus 特别常用的一个函数，比如经常把某个服务的 P99 响应时间来衡量服务质量。不过它到底是什么意思很难解释得清，特别是面向非技术的同学。另一方面，即使是资深的研发同学，在排查问题的时候也经常会发现 `histogram_quantile` 的数值出现一些反直觉的“异常现象”然后摸不着头脑。本文将结合原理和一些案例来分析这个问题。

## 统计学含义

Quantile 在统计学里面中文叫[分位数](https://zh.wikipedia.org/wiki/%E5%88%86%E4%BD%8D%E6%95%B0)，其中 X 分位数就是指用 X-1 个分割点把概率分布划分成 X 个具有相同概率的连续区间。常用的比如有二分位数，就是把数据分成两个等量的区间，这其实就是[中位数](https://zh.wikipedia.org/wiki/%E4%B8%AD%E4%BD%8D%E6%95%B8)了。还有当 X=100 时也叫[百分位数（percentile）](https://zh.wikipedia.org/wiki/%E7%99%BE%E5%88%86%E4%BD%8D%E6%95%B0)，比如我们常说 P95 响应延迟是 100ms，实际上是指对于收集到的所有响应延迟，有 5% 的请求大于 100ms，95% 的请求小于 100ms。

Prometheus 里面的 `histogram_quantile` 函数接收的是 0-1 之间的小数，将这个小数乘以 100 就能很容易得到对应的百分位数，比如 0.95 就对应着 P95，而且还可以高于百分位数的精度，比如 0.9999。

## quantile 的“反直觉案例”

**问题1：P99  可能比平均值小吗？**

正如中位数可能比平均数大也可能比平均数小，P99 比平均值小也是完全有可能的。通常情况下 P99 几乎总是比平均值要大的，但是如果数据分布比较极端，最大的 1% 可能大得离谱从而拉高了平均值。一种可能的例子：

```
1, 1, ... 1, 901 // 共 100 条数据，平均值=10，P99=1
```

**问题2：服务 X 由顺序的 A，B 两个步骤完成，其中 X 的 P99 耗时 100ms，A 过程 P99 耗时 50ms，那么推测 B 过程的 P99 耗时情况是？**

直觉上来看，因为有 X=A+B，所以答案可能是 50ms，或者至少应该要小于 50ms。实际上 B 是可以大于 50ms 的，只要 A 和 B 最大的 1% 不恰好遇到，B 完全可以有很大的 P99：

```
A = 1, 1, ... 1,  1,  1,  50,  50 // 共 100 条数据，P99=50
B = 1, 1, ... 1,  1,  1,  99,  99 // 共 100 条数据，P99=99
X = 2, 2, ... 2, 51, 51, 100, 100 // 共 100 条数据，P99=100
```

如果让 A 过程最大的 1% 接近 100ms，我们也能构造出 P99 很小的 B：

```
A = 50, 50, ... 50,  50,  99 // 共 100 条数据，P99=50
B =  1,  1, ...  1,   1,  50 // 共 100 条数据，P99=1
X = 51, 51, ... 51, 100, 100 // 共 100 条数据，P99=100
```

所以我们从题目唯一能确定的只有 B 的 P99 应该不能超过 100ms，A 的 P99 耗时 50ms 这个条件其实没啥用。

**问题3：服务 X 由顺序的 A，B 两个步骤完成，其中 A 过程 P99 耗时 100ms，B 过程 P99 耗时 50ms，那么推测服务 X 的 P99 耗时情况是？**

有人觉得答案是“不超过 150ms”，理由是 A 过程的 P99 是 100ms，说明 A 过程只有 1% 的请求耗时大于 100ms，同理 B 过程也只有 1% 的请求耗时大于 50ms，当这两个 1% 恰好撞上才会产生 150ms 的总耗时，绝大多数情况下总耗时都是小于 150ms 的。

此处问题同样在于认为数据是“常规分布”的，假如 A 过程和 B 过程最大的 1% 大的离谱，例如都是 500ms+，那么服务 X 就会有 1%-2% 的请求耗时 500ms+，也就是说服务 X 的 P99 耗时会在 500ms 以上：

```
A = 1, 1, ...  1,   1, 100, 500 // 共 100 条数据，P99=100
B = 1, 1, ...  1,   1,  50, 500 // 共 100 条数据，P99=50
X = 2, 2, ... 51, 101, 501, 501 // 共 100 条数据，P99=501
```

**问题4：服务 X 有两种可能的执行路径 A，B，其中 A 路径统计 P99 耗时为 100ms，B 路径统计 P99 耗时 50ms，那么推测服务 X 的 P99 耗时情况是？**

这个问题看上去十分简单，如果所有请求都走 A 路径，P99 就是 100ms，如果都走 B 路径的话，P99 就是 50ms，然后如果一部分走 A 一部分走 B，那 P99 就应该是在 50ms ~ 100ms 之间。

那么实际上真的是这样吗？我经过仔细的研究，最后发现确实就是这样的……乍看上去这个问题跟涉及到平均数的[辛普森悖论](https://zh.wikipedia.org/wiki/%E8%BE%9B%E6%99%AE%E6%A3%AE%E6%82%96%E8%AE%BA)有些像，似乎可以通过调整 A 路径和 B 路径的比例搞出一些幺蛾子，但其实不论 A 跟 B 是怎样的比例，从数量上看，大于 100ms 的请求最多只有 1%A + 1%B = 1%X 个，因此 X 的 P99 不会大于 100ms，同理小于 50ms 的请求不会多于 99%X 个，可知 X 的 P99 也不会小于 50ms。

**问题5：某服务 X 保存数据的过程是把数据发给数据库中间件 M，中间件 M 有 batch 机制，会把若干条并发的请求合并成一个请求发往数据库进行存盘。如果测得 X 保存数据耗时的 P99 为 100ms，那么推测 M 请求数据库的 P99 耗时情况是？**

关键点在于一个请求的多个步骤不是一一对应的，这种情况在分布式系统中并不罕见，我们需要具体情况具体分析，很难简单地推断 M 的 P99 耗时。

最容易注意到的，M 的高延迟能在多大程度上影响 X 的延迟，跟 batch size 息息相关。例如 M 存在一些耗时特别高请求，但是对应的 batch size 恰好很小，这样对 X 的影响就比较有限了，我们就可能观察到 M 的 P99 远大于 X 的 P99 的现象。与之相反，如果对应的 batch size 恰好特别大，极少量的 M 高延迟也会体现在 X 的统计中，我们就能观察到 X 的 P99 远大于 M 的 P99 的现象。

再比如 M 在连接数据库时可能使用了连接池，如果少量的数据库请求过慢，可能导致连接池发生阻塞影响后续的大量存盘请求，这时 M 统计到的高延迟请求很少，而 X 统计到的高延迟会很多，最终也能形成 X 的 P99 远大于 M 的 P99 的状况。

## histogram 场景下的 quantile

前面的内容都是从 quantile 的定义出发的，并不限于 Prometheus 平台。具体针对 Prometheus 里的 `histogram_quantile`，还有一些要注意的点。

一个是因为 histogram 并不记录所有数据，只记录每个 bucket 下的 count 和 sum。如果 bucket 设置的不合理，会产生不符合预期的 quantile 结果。比如最大 bucket 设置的过小，实际上有大量的数据超出最大 bucket 的范围，最后统计 quantile 也只会得到最大 bucket 的值。因此如果观察到 `histogram_quantile` 曲线是笔直的水平线，很可能就是 bucket 设置不合理了。

另一种情况是 bucket 范围过大，绝大多数记录都落在同一个 bucket 里的一段小区间，也会导致较大的偏差。例如 bucket 是 100ms ~ 1000ms，而大部分记录都在 100ms ~ 200ms 之间，计算 P99 会得到接近于 1000ms 的值，这是因为 Prometheus 没记录具体数值，便假定数据在整个 bucket 内均匀分布进行计算。[Prometheus 的官方文档](https://prometheus.io/docs/practices/histograms/#errors-of-quantile-estimation)里也描述了这个问题。

