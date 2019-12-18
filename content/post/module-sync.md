---
title: "游戏逻辑模块组织及数据同步"
date: "2014-09-28"
tags: ["游戏开发"]
description: "易于管理的服务器客户端数据同步方案"
---

一个游戏根据功能可以划分为多个不同的模块，如金钱、背包、装备、技能、任务、成就等。按照软件工程的思想，我们希望分而治之单独实现不同的模块，再将这些模块组合在一起成为一份完整的游戏。但现实是残酷的，不同模块之间往往有千丝万缕的联系，比如购买背包物品会需要扣金币、打一个副本会完成任务，完成任务又会奖励金币和物品，金币的增加又导致一个成就达成。于是我们虽然在不同的类或不同的文件中来实现各个模块，却免不了模块间的交叉引用和互相调用，最后混杂不堪，任何一点小修改都可以导致牵一发而动全身。

为了后面说明方便，我们考虑这样一个小型游戏系统：总共有 3 个模块，分别是金钱、背包、任务。购买背包物品需要消耗金币，卖出背包物品可得到金币，金币增加到一定数额后会导致某个任务的状态变为完成，完成任务可获得物品和金币。这 3 个模块的调用关系如图。

![1](/assets/img/module-sync-1.png)
 

首先我们把模块的数据和逻辑分离，借鉴经典的 MVC 模式，数据部分叫作 Model，逻辑部分叫作 Controller。如此一来，游戏功能部分就被划分出来了两个不同的层次，Controller 处于较高的层次上，可以引用一个或者多个 Model。Model 层专心处理数据，对上层无感知。每个 Model 都是完全独立的模块，不引用任何 Controller 或 Model，不依赖于其他任何对象，可以单拿出来进行单元测试。

![2](/assets/img/module-sync-2.png)

对于我们的例子，每个模块提供的接口列举如下：

BagModel：获取物品数量，增加物品，扣除物品

MoneyModel：获取金币数量，增加金币，扣除金币

TaskModel：增加任务，删除任务，标记任务为完成

BagController：购买物品，卖出物品

TaskController：完成任务

购买或卖出物品时，由 BagController 进行或操作校验，随后调用 BagModel 和 MoneyModel 完成数据修改。完成任务时，由 TaskController 调用各个模块。

现在唯一的问题是，既然 MoneyModel 不引用其他模块，那么在金币增加时如何告知任务模块去完成任务呢？这里我们需要引入一个管理依赖的利器：观察者模式。

具体使用方式是把 Model 实现为一个 Subject，对某个 Model 的数据变化感兴趣的 Controller 实现为对应的 Observer。我们的例子中，MoneyModel 是 Subject，在金币数量变化时通知所有已注册的 Observer；TaskController 是 MoneyModel 的一个 Observer，在初始化时向 MoneyModel 注册。

![3](/assets/img/module-sync-3.png)

注意图中由 MoneyModel 指向 TaskController 的虚线箭头，代表 MoneyModel 数据变化时会去通知 TaskController，用虚线是因为 MoneyModel 并不依赖于 TaskController（只依赖于 Observer 接口）。同样 BagModel 也可以提供背包物品变化的 Subject，如果新加一个任务是要求某物品的数量达某个值，那么 TaskController 可向 BagModel 注册，这样在物品变化时就能得到通知了，图中也画出了这条虚线。

对观察者模式不熟悉的读者朋友可以自行查阅资料， 本文的重点并不是介绍设计模式。这里简单提示一下观察者模式的精髓：当某模块调用其他模块时就产生了依赖，这时可以不直接去调用，而是转而实现一个机制，这个机制就是让其他模块告诉自己他们需要被调用。最后调用的流程没变，变化的是依赖关系。

 

在客户端情况要更复杂一些，实际上加入 UI 后，我们的模块设计就成经典的 MVC，这也是我们为什么把数据模块和逻辑模块分别叫 Model 和 Controller 的原因。

![4](/assets/img/module-sync-4.png)

这里只画出了背包模块。这里的 System API 指与游戏运行平台相关的一些接口，可能是操作系统 API、引擎 API、图形库 API 等等。View 模块和 Model 模块地位相当，只处理显示而不管游戏功能，需要显示的数据都是由 Controller 提供的。对于能输入的 View 同样采用观察者模式，点击等事件发生时通知其他模块（而不是直接调用），注意图中由 BagView 指向 BagController 的虚线箭头。

下面介绍数据同步的设计。

首先对于网络游戏，客户端所展示的数据是服务器传送过来的。当玩家操作导致数据发生变化时，最好也由服务器更新给客户端。曾经接手过一个项目，很多操作的结果都是客户端先算出来的，于是各种逻辑都是服务器和客户端各实现一遍，很容易两边的数据就不一致了，很让人头疼。

所以我们的同步思路是当客户端向服务器发起一个请求时，服务器将所有变化的数据同步给客户端，客户端收到服务器的返回后再更新数据，绝不私自改动数据。在这个指导思想下，我们消息包结构是这样的（以物品卖出举例）：

```proto
message BagItemSellCG {
    optional int32 id = 1;
    opitnoal int32 count = 2;
}

message BagItemSellGC {
    optional int32 result = 1;
    optional Sync sync = 2;
    opitonal BagItemSellCG postback = 3;
}
```

服务器向客户端返回的消息几乎总是包含 3 个字段。result 为操作结果可能是 0 或者错误码，sync 中包含了所有的数据更新，postback 将客户端的请求消息原封不动返回去，便于客户端进行界面更新或友好提示。

sync 是一个比较复杂的 message，包含了所有需要更新的 Model 的数据。感谢 Protocol Buffer 的 optional 选项，大多数情况下我们发送的数据只是其中很小的一部分。

先来看服务器端消息处理和同步的设计。

![5](/assets/img/module-sync-5.png)

如图所示，我们在 Model 和 Controller 之上新加了一个 Handler 接口层。Handler 负责解析消息包，调用 Controller 处理消息包，在必要的时候调用 SyncController 构建同步数据，最后打包成消息返回给客户端。

每个 Model 在管理数据的基础上会维护变化数据的集合，对于简单的 Model 比如 MoneyModel 就是一个 bool 脏标记，而 BagModel 则维护变化物品 id 的集合。变化数据列表在同步之后清除。

 

客户端的结构是类似的。

![6](/assets/img/module-sync-6.png)

与服务器的区别就在于 SyncController 是负责调用 Model 更新数据，每个 Model 都实现数据更新接口。注意除 SyncController 之外，其他 Controller 只能读取 Model 而不能改变其数据，这样就保证了所有数据一定是从服务器同步的。

最后我想以出售物品为例子完整走一遍流程。从客户端进行操作开始，到请求发到服务器，最后再返回客户端更新数据和界面。完整的图比较复杂，混在一起基本上没法看了，只好删掉了客户端的任务模块……

![7](/assets/img/module-sync-7.png)

1. BagView 界面产生一个点击，因为 BagController 是 BagView 的观察者，所以 BagController 能得到点击事件的通知。
2. BagController 识别出此点击是要出售物品，于是构建好消息包发往服务器。
3. 服务器识别出消息类型是 Sell，于是消息被派发给 SellHandler。
4. SellHandler 调用 BagController 执行逻辑。
5. BagController 取出 BagModel 和 MoneyModel 的数据进行条件检查，如果无法执行操作则生成错误码返回给 SellHandler，否则调用 Model 修改数据，此时 BagModel 会记录下变化物品的 id，MoneyModel 会做一个脏标记。
6. MoneyModel 数据发生变化，通知自己的观察者（TaskController）。
7. TaskController 判断任务完成，调用 TaskModel 更新数据。TaskModel 会记录发生变化的任务。
8. SellHandler 对 BagController 的调用返回后，如果出错则直接返回消息包给客户端。否则调用 SyncController 收集同步数据。
9. SyncController 调用各个模块收集同步数据，各个模块提交同步数据后清除自己维护的标记。
10. SellHandler 将操作结果和同步数据打包后发往客户端。
11. 客户端识别出消息类型是 Sell，消息被派发给 SellHandler。
12. BagHandler 将消息处理结果发给 BagController。
13. BagController 根据消息处理结果，通知 BagView 进行必要的提示。
14. SellHandler 将消息包中的数据同步部分发给 SyncController。
15. SyncController 将同步数据同步给各个模块。
16. BagModel 和 MoneyModel 的数据发生了变化，通知观察者，即对应的Controller。
17. Controller 调用View  进行界面更新。

-----------------------

## Q&A

### 返回客户端提交的 postback 对于网络传输来说太过重量级, 可以尝试改为客户端保存一个 rid-postback 的键值对, id 由客户端自增, 请求数据时把 rid 一起发送给服务器

支持你的方案。

但我的想法不是出于数据量的考虑，因为一般网游客户端发往服务器的消息都是比较小的，服务器返回的消息会比较大。
原因是后来我们考虑到消息可能丢包的问题，当丢包发生时，客户端需要重发请求，这样一来 rid 检验及保存之前发送的请求就是必须的了。而保存下来的请求正好又可以用来替代上文的 postback ，所以你的方案非常合理。

### 我使用了背包里一个物品,在返回的 sync 中是返回使用掉的物品信息, 还是背包的全部物品信息?

因为我们背包里的物品会比较多，所以同步全部物品是不合适的。

我们的做法是删除物品后记录物品 id，生成同步数据时如果发现对应 id 的物品不存在，则同步一个数量为 0 的物品信息，客户端收到数量为 0 的物品后做删除操作。
有的模块没有一个代表删除的特殊“零值”，比如任务。我们的做法是将新增/更新与删除分开同步：

```proto
message TaskSync {
	repeated Task update = 1;
	repeated int32 delete = 2;
}
```
