---
title: "重复的代码都应该被消除吗？"
date: 2020-07-22
tags: ["杂谈"]
---

昨天，我们项目组同事给测试代码做了些重构和梳理，总体上是很好的，不过有一个地方我觉得其实不如原本的好。

这段代码是在单测里虚拟一个集群添加一些节点，原来的代码是这样的（无关细节略作调整）：

```go
testCluster := mockcluster.NewCluster(opt)
testCluster.AddLabelsStore(1, 1, map[string]string{"zone": "z1"})
testCluster.AddLabelsStore(2, 1, map[string]string{"zone": "z1"})
testCluster.AddLabelsStore(3, 1, map[string]string{"zone": "z2"})
testCluster.AddLabelsStore(4, 1, map[string]string{"zone": "z2"})
testCluster.AddLabelsStore(5, 1, map[string]string{"zone": "z3"})
```

优化过之后的代码是这样的：

```go
testCluster := mockcluster.NewCluster(opt)
allStores := []struct {
    storeID     uint64
    regionCount int
    labels      map[string]string
}{
    {1, 1, map[string]string{"zone": "z1"}},
    {2, 1, map[string]string{"zone": "z1"}},
    {3, 1, map[string]string{"zone": "z2"}},
    {4, 1, map[string]string{"zone": "z2"}},
    {5, 1, map[string]string{"zone": "z3"}},
}
for _, store := range allStores {
    testCluster.AddLabelsStore(store.storeID, store.regionCount, store.labels)
}
```

令我比较诧异的是，团队里面有一半以上的人都觉得新版本更好，这让我觉得有必要写写这个问题好好整理一下我的理由。

认为新版更好的理由主要有 2 点：

1. 旧版本重复了多次 `AddLabelStore` 这个方法调用，新版本不存在重复代码，更加优雅。
2. 新版本对修改更加友好，如果测试逻辑变成调用 `AddLabelStore` 之后要加入其他逻辑，新版只用改一个地方。

-----

我先解释一下为什么这两点理由是不成立的。

### 重复代码未必糟糕

大家都听说过 DRY（Don't Repeat Yourself）原则，不过很少有人知道这个原则原本并不是针对重复代码。

> Every piece of knowledge must have a single, unambiguous, authoritative representation within a system.

注意这里说的是 knowlege 而不是 code。有几个层面的区别：

其一，代码可能并不形成知识。比如上面的例子，重复的部分只是函数调用而已，而函数调用传的参数，每次调用都各不相同，即使使用新版本的方法进行抽象，这些参数还是以一个列表的形式列举了出来，并没有通过整理形成任何公共的知识。

其二，重复的代码可能是不同的知识。[DRY is about Knowlege](https://verraes.net/2014/08/dry-is-about-knowledge/) 一文举了一个很好的例子：两件商品都限购 3 份，但这更多只是巧合，本质上两个不同的知识，因而不能去消除重复代码。

### 提前设计的弊端

第 2 个理由还是比较有道理的，可惜只有在未来真的需要做这样的修改时才成立。如果未来并不会产生如预期一样的修改，甚至过了一段时间整个测试用例都可能被删掉了，那么我们就是白白付出了精力做这个重构。

更重要的是，新版本的代码更不易读（后面分析），每次被阅读时，都会多消耗一份心智负担。即使最终真的发现要如预期般修改了，在此之前所付出的额外精力也是浪费的。

另外，未来如果真要在每个 `AddLabelStore` 后面加几行别的逻辑，未必没有其他更好做法。比如直接在 `AddLabelStore` 之内修改，或者提取一个新函数把 `AddLabelStore` 包装一下，完全没有必要现在就提前做决定。

我一直有一个观点，好的代码应该“活在当下”。既不应该为过去的错误长期买单，也不应该为未来的需要提前透支。

-----

### KISS

KISS（Keep It Simple, Stupid）不仅适用于软件开发，是整个设计领域都通用的经验原则。简单地说就是应该注重简约，避免引入不必要的复杂度。

新版本的代码复杂在什么地方呢？

最简单地判断依据是，新版本的代码更长。或者严谨一点说，新版本代码在没有显著降低单行代码理解成本的前提下，完全相同的功能使用了更多代码，因而阅读起来要付出更多的精力。

我们还可以这么来看。尝试把两份代码试着用自然语言来描述，会发现后者会长得多。为了干同样的事情，新版本引入了一个匿名结构体，一个数组，一个 for 循环，这显然会给读者带来理解上的心智负担，而且在我看来很难说是优雅。

反观旧版代码，没使用精巧的结构和复杂的设计，但是清晰地表达了意图，一眼看上去就知道在干什么，而且显然没什么隐藏的问题。我认为是更好的选择。

