---
title: "自动给 Gmail 中 GitHub 的邮件打标签"
date: 2020-01-15
tags: ["教程", "App Script"]
draft: false
---

相信不少人跟我一样，平时是把 Gmail 直接当成 TODO List 来用的。处理 GitHub Issue 或 PR 时也是基于 Gmail 来完成。根据 PR 和 Issue 的不同状态，给邮件打上对应的标签，是为了能在不点开邮件 thread 的情况下就能评估优先级和快速进行一些处理，效果如下：

![preview](/assets/img/gmail-tag.png)

那么这个是怎么做的呢？我们知道 Gmail 有过滤器的功能可以自动加标签，可惜的是过滤器不支持正则表达式什么的，在 GitHub 邮件这个场景下很容易误判。最后的方法还是祭出了 Google App Script 大法，代码如下：

{{<gist disksing 5eb3b9740cf1921e044fed4b58b08d35>}}

用法是创建一个 Google App Script 项目，把代码贴进去，然后部署为网络应用并授权，最后再加上触发器定时运行就搞定了。注意触发器别设得太频繁了，15 分钟运行一次就差不多了，太频繁可能会超出配额。具体操作流程可以参考下[追踪 GitHub PR review 记录](/review-recorder)一文。

