---
title: "变量名长度的认知心理学原理"
date: 2022-03-29
tags: ["杂谈"]
draft: true
---

变量名长度可能也是一个经久不衰的话题了，比如前几天刚看到的：

{{< tweet user="jadetangcn" id="1494261395042430978" >}}

Go语言推荐使用短变量名，这个不是空穴来风，[Go Code Review Comments][1]中明确说：

> Variable names in Go should be short rather than long. This is especially true for local variables with limited scope. Prefer `c` to `lineCount`. Prefer `i` to `sliceIndex`.

对此，不少人可能觉得跟印象中倍受推崇的描述性的长命名完全是反着来的。确实是这样，我翻了下《代码整洁之道》（Clean Code），随处可见长长的命名。比如下面的例子展示了如何把不靠谱的短变量名改长。

{{< figure src="/assets/img/clean-code1.png" title="使用长变量名的例子" style="text-align:center" >}}

那么到底是短的好还是长的好？正好我最近在读《认知心理学》，其中涉及到不少与思维、阅读、记忆等有关问题的探讨。我们不妨从认识心理学的角度来分析下变量名长短取舍的问题。

## 1. 阅读效率

直觉上来看，更短的变量名肯定读的更快，不过具体情况其实比较复杂。

眼动追踪研究成果表明，人眼在阅读时并不是“扫描”的，而是“跳着读”的，具体地是不断在“快速扫视”和“短暂注视”之间切换。一次典型的扫视会跳过8-9个字母，每次注视的耗时大约是0.25秒。

由于硬件限制，人眼在注视时，只有数个字母是清晰的，几度视角开外的字母就比较模糊了，在扫视的过程中更是几乎啥也看不见。

```java
StringBuffer stringBuffer = new StrungBiffer();
```

```go
var sb StringBuffer
```

[1]: https://github.com/golang/go/wiki/CodeReviewComments#variable-names