---
title: "大功能怎么拆分PR"
date: 2022-06-29
tags: ["杂谈"]
draft: true
---

主干开发模式（trunk based development）是当下流行的开发模式，它的主要好处是可以避免维护大量的活跃分支，并且对CICD（持续集成/持续交付）更加友好。

与主干开发模式分庭抗礼的另一种模式是功能驱动开发模式（feature driven development），它的做法是为每个开发中的功能长期维护分支，直到开发测试完成后再合进主干。

功能驱动开发模式的拥趸常常以“频繁冲突”和“主干不稳定”为理由来诟病主干开发模式。本文我不想来评判二者孰优孰劣，而是想聊聊在主干开发模式下，应该如何把大功能拆解成小PR，以及对上述两点“弊端”的看法。

### 拆PR的时机

很多人开发大功能的流程是在本地维护一个分支，先在这个分支上“调通”整个功能，然后再把这个大分支拆分成小PR，顺带补充单元测试，依次合进master。

不得不说，这么操作确实很大程度上避免了“频繁冲突”和“破坏主干稳定性”的难题。但问题在于这不是主干开发模式，在开始拆PR合并之前仍然是在分支上做功能驱动开发，只不过功能分支维护在开发者本地，而且这个本地分支长时间没人review，没有单元测试，我觉得还不如正规的功能驱动开发模式。

要应用主干开发模式，首先就要破除心理障碍，不要怕主干上有残缺不全的功能，也不要怕开发功能过程中走的“弯路”提交进主干“污染”了主干的提交历史。这些都是爽快享受主干开发的必要代价，真正需要特别关心的是避免开发中的功能影响原有功能的正常运行。

因此，最好的拆PR的时机其实是在写代码之前，每次只完成一小步，写完之后就提交PR合进master，这样才是真正的“在主干上开发”。
