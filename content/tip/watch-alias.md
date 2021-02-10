---
title: "watch alias 不能工作的问题"
date: 2020-11-20
tags: ["命令行"]
---

watch 是一个常用的命令，不过它与 alias 不能很好地工作，比如：

```
$ watch ll
sh: 1: ll: not found
```

其实只需要把 watch 也 alias 一下加个空格就行了：
``
$ alias watch='watch '
$ watch ll
``
