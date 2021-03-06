---
title: "protobuf 改字段类型"
date: 2020-08-07
tags: ["protobuf"]
---

我们知道 protobuf 在序列化的时候，是使用 filed number 来标识不同字段的，因此在发布以后仍然可以修改字段名，同时保证兼容。

那么，字段的类型能不能修改呢？其实也是可以的，只要两种类型的 wire type 是相同的，那么就可以互相兼容，比如 int32, int64, uint32, bool, enum 互相都兼容。详见[官方文档](https://developers.google.com/protocol-buffers/docs/proto3#updating)。

具体有啥用呢？

比如我们一开始有个字段只能取 true, false 两种值，很自然地，我们会用 bool。之后发现其实还有 unknown 这种情况，这时就不用再加一个字段，直接定义一个这样的 enum：

```
enum State {
    False = 0;
    True = 1;
    Unknown = 2;
}
```

但是要注意的是，这时兼容是单向的，即新版本能兼容旧版本，但是旧版本不能兼容新版本。例如新版本写了个 Unknown 发给旧版，那肯定是读不出来的。
