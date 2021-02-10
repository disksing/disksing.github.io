---
title: "如何解压 gz 类型的文件"
date: 2020-07-21
tags: ["命令行工具"]
---

一般在 Linux 平台，我们见到的压缩文件类型都是 .tar 或者 .tar.gz，解压方式分别是 `tar -xf` 和 `tar -zxf`，也就是当后缀多个 `.gz` 时要多加一个 `-z` 来解压。

还有一种不太常见的情况，就是文件没有 `.tar` 而是直接 `foo.gz`，这时需要使用 gzip 命令来解压 `gzip -d foo.gz`。