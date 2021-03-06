---
title: "战斗模块设计"
date: "2014-10-15"
tags: ["游戏开发"]
description: "卡牌游戏支持多种战斗模式的困局"
---

本文所描述的战斗模块设计方案源自于实际项目（卡牌手游）的需求，可能并不适用于 MMORPG、ARPG 等类型游戏。本文所涉及战斗的基本形态是：从游戏环境中收集战斗所需要的数据，随后在一个独立封闭的环境中进行若干次迭代计算（其间可能会读入玩家的指令输入），最后达成某些条件后分出胜负战斗结束。如果感觉这个描述太抽象，可以参考一下三国志系列的战斗或者 QQ 斗地主 :P

本文重点分析战斗系统在模块这个级别上的设计和取舍，不涉及具体游戏的战斗计算逻辑。

需求分析
--------------

项目战斗采用的是自动回合制战斗，即势力双方派出若干角色按照特定的队形排好后依次进行攻击，直到其中一方所有角色都阵亡则战斗过程结束。为了提高战斗的趣味性和策略性，我们加入了手动释放技能作为补充，即玩家可主动选择释放技能的时机（参考刀塔传奇）。

从程序的角度来看，如果战斗过程全自动不需要玩家干预，那么可以用“秒算”的方式来做。也就是直接一个循环瞬间计算出完整的战斗过程，战斗过程保存下来后再慢慢播放。一旦引入玩家的手动操作，秒算就不再凑效了，因为战斗过程的计算会受实际操作的影响。

按照策划的设计，目前游戏中战斗的类型可以分为 3 种。副本（PvE）：由玩家势力和配置的 NPC 势力对战，玩家可以操作技能释放；排位赛（PvP）：由两方玩家事先派出的势力对战，因为玩家不一定在线，所以技能是自动释放的，不再支持手动操作；挑战赛（PvP）：当两方玩家都在线时发起对战，这时双方都可以操作技能释放。其中

1. 副本战斗要求在网络连接不稳定时也能正常进行，需要由客户端进行战斗计算，等到结束后再将结果上传到服务器验证。
2. 排位赛存在玩家不在线的可能，所以是服务器进行“秒算”，并保存战斗录像供之后播放。
3. 挑战赛在服务器计算，在战斗过程中客户端需要同步操作至服务器。

对于第 1 点需求我是存在异议的，为了网络不稳定情况下的体验，导致游戏中最复杂的战斗部分需要服务器和客户端各实现一遍，工作量增加不说，最重要的模块的复杂程度急剧上升，需要同时支持三种模式：

1. 客户端计算，服务器验证。
2. 服务器秒算，客户端播放。
3. 服务器计算，客户端播放过程和读取操作，并实时同步。

此外战斗逻辑本身是非常复杂的，几乎一定会出 bug 并且不容易测试、发现 bug 也不容易重现、重现了也不容易调试和修正，服务器和客户端还要各自实现一遍导致这一系列成本成倍增加。（项目中服务器和客户端编程语言不一致无法共用代码）

战斗模块的设计，重点就在于为这两个棘手的问题提供一个解决方案。

从外部看战斗模块
---------------

从外部看战斗模块，就是规划模块的外部接口，划定模块与外部系统的界线。

用极简的视角来看战斗模块，可以认为它就是一个函数，输入是战斗需要的所有数据（包括双方势力的战斗单位布局，每个战斗单位的出手速度、攻防血、技能等属性），输出是战斗结果（胜负情况，可能还包括战斗过程记录）。

战斗模块不关心的是：战斗从哪里触发；战斗结束后要更新哪些数据；战斗势力是玩家还是NPC；战斗能否发生，如玩家是否有足够的体力，玩家等级是否满足副本等级，是否领取了对应的任务等。

战斗模块关心的是：战斗过程的迭代，战斗结果的判定，战斗过程中数据的网络同步（如果需要），战斗过程的展示（客户端）。

特别注意战斗作为一个独立模块不应该有任何外部依赖。例如某单位的攻击力是由配表中的数值加上其等级进行计算，再综合各种加成得出的，那么应当是在战斗模块外部算出最终数值后再交给战斗模块，而不应该由战斗模块去调用外部接口进行计算。这是很自然的，因为我们一定不想由于某张配表变化或者某模块数据结构的调整导致需要修改战斗模块的代码，最后引入 bug 带来不必要的麻烦。

从内部看战斗模块
------------------

从内部看战斗模块，也就是制定模块的实现方法，重点是要同时支持3种战斗模式。

思考问题的过程其实特别快，有时候想法的产生来自于直觉没有什么特别的理由，所以这里只介绍最后想出来的方案……

战斗模块从内部划分为这么几个组件：数据，计算，展示（仅客户端），输入。

* 数据部分是从模块外部传入的，一部分数据会交给计算模块进行迭代（战斗角色的排布和属性等），一部分数据会交给展示组件用于界面展示（双方名字等级等）。
* 计算组件负责战斗迭代，它的输出是一系列战斗过程。
* 展示组件接收战斗过程，并在场景中展示出对应的模型、动画、UI。
* 输入组件负责读取用户输入。

接下来我们看看这些组件如何组合起来满足 3 种战斗模式。

1. 副本：副本战斗完全在客户端运行，计算组件输出的战斗过程发往展示组件，输入组件读取到的操作发往计算组件影响后续迭代。
2. 排位赛：排位赛由服务器秒算，服务器直接计算出所有战斗过程后返回给模块外部存储下来。客户端查看战斗记录时，计算好的战斗过程以数据的形式发往客户端战斗模块，此时客户端的计算组件退化为“播放器”，只需要将服务器生成的战斗过程依次发往展示组件。
3. 竞技场：竞技场模式操作和战斗过程都是通过网络实时同步的。服务器这边：输入由客户端通过网络发送过来，计算出的战斗过程通过网络发往客户端的展示组件。客户端这边：展示组件从网络接收战斗过程，输入组件读取到输入后发往服务器。

战斗过程的记录方式
-----------------

战斗过程同时用于组件间通信和战斗录像的保存，其数据结构有必要探讨一下。
首先记录方式应该是基于“打谱”而不是“快照”。两者的区别是这样的，打谱类似于“回合 1：A 攻击 B 产生 10 点伤害；回合 2：B 攻击 A 产生 5 点伤害”，而快照类似于“回合 1：A 100 生命 B 90 生命；回合2： A 95 生命 B 90 生命”。基于打谱的原因有二：其一是通常打谱产生数据量要远小于快照，不妨想一下象棋的棋谱，几十分钟的一局对弈用棋谱记下来不过半页纸；其二是基于事件的记录形式更便于展示组件展现过程。

但是打谱的记录方式很容易夹带一些隐晦的问题，究其原因战斗播放是依据“谱”的一个复现过程，播放到任意时刻的状态是由前面一系列的步骤推演出来的，只要有一步出现偏差最后的结果就可能大相径庭。为了保证一致，我们的战斗过程记录一定要足够直接。例如受到的伤害值减掉防御值得出 HP 的损耗，就必须直接记录 HP 的损耗，而不能只记录伤害值交给播放组件来进行计算。更好的做法是将打谱法和快照法结合着使用，同时记录 HP 的损耗和最后剩余的 HP，这样即使不慎出现了不一致也能很快恢复。

可能有同学不理解，损耗=伤害-防御，如此一个简单的计算怎么会发生不一致呢？可能还有同学会觉得自动战斗根本没必要记录战斗过程，因为没有操作的影响，直接拿初始数据重新推演一下不就出来了？这是新人常见的思维漏洞，他们忽略了一个重要因素，就是线上网络游戏是在不断演化的。在今天损耗=伤害-防御，下个版本可能就变成损耗=攻击-防御，录像数据还是原来的数据，于是版本一更新战斗录像的过程就全变了，这可就太坑爹了。网络游戏迭代更新很快，一定要时刻提防着数据兼容的问题。

测试的困境
--------------------

前面还提到了另外一个棘手的问题，复杂的战斗逻辑被服务器和客户端各实现一遍，给测试带来了不小的压力。其实换个思维方式就能很巧妙的解决，一旦想通后甚至有一种“塞翁失马，焉知非福”的感觉！

首先要注意不管是服务器还是客户端，战斗计算组件的输出都是一样的：一份完整的战斗过程记录。那么如果两边都正确实现功能的话，给这两个计算组件以相同的输入，则一定可以得到相同的输出（这里不考虑计算过程中的随机数，或者可以认为随机数种子当作参数传入）；反之如果对于相同的输入得到了不同的输出，那就说明至少有一方的实现是有问题的。

基于这个思路，我们可以把两边的战斗模块单提出来独立编译（因为战斗模块的实现不依赖于其他模块，单提出来是很容易做到的），再写个测试程序不断随机生成战斗初始数据分别发往两份实现，收回两边的输出后进行校对，这样容易就能发现 bug 了。等到两份实现能一致地处理大量随机数据时，我们基本上就可以认为两份实现都是正确的了。毕竟两个程序员分别使用不同的编程语言，很巧合地设计出了相似的代码结构，并更加巧合地犯了同一个错误，这个概率应该是足够小的。

-----------------------

## Q&A

### 计算中涉及浮点数时可能导致服务和客户端不一致或是精度误差，如何解决？

游戏实际运行时，不管是哪种战斗模式，负责数值计算的都只是服务器和客户端其中一方，另一方只负责接收结果，所以不怕有精度误差。

对于客户端计算战斗，服务器验证的情况，我们不准备严格核对战斗过程，只准备划定阀值做简单验证。（因为客户端没做自动更新，会有多个版本客户端同时运行的情况，严格验证逻辑是很难行通的）

对于文中最后一部分提到的对比测试的情形。说实话，发文时我完全没想到精度误差这个问题- -#，感谢提醒。我觉得要让不同语言计算结果保持一致，需要在文档中详细规定公式中数值的类型，以及浮点数取整的时机。比如所有数值计算都用双精度浮点数，直到最后得出伤害值时再向下取整为整数。正常情况下不同语言的浮点数计算应该都是依照 IEEE754 实现，所以理论上说是可以得到一致结果的……

