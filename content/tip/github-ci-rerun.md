---
title: "重新触发 GitHub 运行 CI"
date: 2020-08-07
tags: ["GitHub"]
---

有时候，CI 因为莫名其妙的原因卡住，导致 PR 被卡住不能合，提供几种解决思路：

- 点 detail 链接进入对应 CI（比如 travis）的详细界面，一般都有重新运行的按钮可以点
- 创建空提交并 push（使用 git commit --allow-empty）
- 关闭 PR 再 reopen

