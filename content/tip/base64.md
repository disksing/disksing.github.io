---
title: "base64 中的换行符"
date: 2020-07-16
tags: ["编码", "RFC", "命令行工具"]
---

这是 wikipedia 在 [Base64](https://zh.wikipedia.org/wiki/Base64) 词条给出的一个例子：

```
TWFuIGlzIGRpc3Rpbmd1aXNoZWQsIG5vdCBvbmx5IGJ5IGhpcyByZWFzb24sIGJ1dCBieSB0aGlz
IHNpbmd1bGFyIHBhc3Npb24gZnJvbSBvdGhlciBhbmltYWxzLCB3aGljaCBpcyBhIGx1c3Qgb2Yg
dGhlIG1pbmQsIHRoYXQgYnkgYSBwZXJzZXZlcmFuY2Ugb2YgZGVsaWdodCBpbiB0aGUgY29udGlu
dWVkIGFuZCBpbmRlZmF0aWdhYmxlIGdlbmVyYXRpb24gb2Yga25vd2xlZGdlLCBleGNlZWRzIHRo
ZSBzaG9ydCB2ZWhlbWVuY2Ugb2YgYW55IGNhcm5hbCBwbGVhc3VyZS4=
```

比较有趣的是，其中的换行并不是 wrap 产生的，而是实实在在的换行符。在 [RFC2045 (1996)](https://www.ietf.org/rfc/rfc2045.txt)中，规定了每行最多不能超过 76 字符，当消息过长时需要加入换行符。不过，RFC2045 的主要目的不是规范 Base64，而是针对 MIME 的。后来，在针对 Base64 的 [RFC3548](https://tools.ietf.org/html/rfc3548) 和 [RFC4648](https://tools.ietf.org/html/rfc4648) 中都明确规定了不要再插入换行符。

因为这个历史原因，很多工具，包括 linux 常用的 base64 命令行工具，都保留了 76 字符后换个行的行为。如果不想要换行符，可以通过 `-w 0` 参数来取消。

```
echo -n "apfjxkic-omyuobwd339805ak:60a06cd2ddfad610b9490d359d605407" | base64 -w 0
```

## 参考资料
- [stackoverflow](https://superuser.com/questions/1225134/why-does-the-base64-of-a-string-contain-n)