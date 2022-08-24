---
title: "给TiDB（MySQL）写一个代理网关"
date: "2022-08-24"
tags: ["TiDB", "MySQL", "serverless"]
description: "引入数据库网关来优化TiDB Cloud服务运营成本的故事，以及处理MySQL协议的糟心细节"
---

转到cloud团队（主要做TiDB Cloud DevTier）后，这几个月大部分时间都在tidb-gateway这么个项目上折腾。现在一期功能算是上线了，准备开始做二期，趁这个机会简单总结一下。

因为TiDB是兼容MySQL协议的，所以主要其实就是折腾MySQL协议，然后如果你想做一个MySQL Gateway，大部分内容应该也是兼容的。我把相关代码整理了一下放在 [tidb-gateway](https://github.com/oh-my-tidb/tidb-gateway)项目。

## 项目背景

先简单说下为什么需要做网关。由于TiDB Cloud是没有多租户或者serverless支持的，这就是说，用户每创建一个集群（包括免费的DevTier），我们在后台就会真给创建一个独立的集群。

跟友商的serverless方案相比，我们这么做的好处大概就是开发速度快，不用为上云做特别的改动，然后缺点就是贵。

为了降低成本，我们在DevTier上做了一定的“体验降级”：当用户的集群在连续一段时间不使用之后，我们会保存数据并把集群休眠，下次用户再需要使用的时候，需要先进行一个手动唤醒操作。

这么做完了有一定效果，至少对于已经“跑路”的集群，我们不用无限期地付出成本了。但是要继续优化，我们遇到两个障碍：

1. 由于MySQL协议的限制，客户端通过TCP连接到服务器后，是由服务器首先发送第一个消息，同时因为是裸TCP连接，不像HTTP有请求Header可以知道客户访问的域名来进行路由。这样我们不得不为每个用户集群创建独立公网LB——据说这个还挺贵的。

2.  临时关停的集群需要在网站上手动开启，用户体验负分，这导致我们权衡之下只针对静默7天以上（基本判定是跑路了）的集群做休眠处理，这样对每个集群都额外付出了7天的成本。

解决这两个问题的方法自然就是在TiDB前端引入一个网关服务了。

网关负责接受客户端连接并与之交换消息，等拿到用户信息之后，以代理的方式去连接真正的用户集群。同时，如果用户集群处于休眠状态，网关可以把连接阻塞，然后通知K8s唤醒，这样一来用户在休眠后第一次连接时等待一段时间，不需要在网页端做额外操作了。

## MySQL建立连接过程

我们先简单分析一下MySQL的连接建立流程。

1. 客户端向服务端建立TCP连接。

2. 服务端返回`InitialHandshake`消息，其中包括版本号和一些兼容性标记（比如是否支持TLS等）

3. 客户端返回`HandshakeResponse`消息，其中包括兼容性标记、连接使用的用户名及数据库名。

4. 服务端和客户端根据`AuthMethod`交换若干次消息，直到服务端返回`Ok`或者`Err`消息，说明连接成功建立或者失败。

### TLS连接建立过程

如果需要启用安全连接，步骤3中，Client会先发送半个`HandshakeResponse`消息包，其中携带了`ClientSSL`标记，服务端读到此标记后，会发起将TCP连接升级为TLS连接，升级完成后，Client会再次发送`HandshakeResponse`消息回归到常规流程。

### 鉴权FastPath及AuthMethod磋商

为了减少建立连接过程种消息交换的次数，MySQL Protocol有一个鉴权的快速通道。

在服务端发送`InitialHandshake`消息时，会先默认猜一个`AuthMethod`，并随机生成8字节或者更长的`challenge payload`，放在`InitialHandshake`消息中一起发给客户端。（为什么说是“猜”呢，因为不同用户可能设置不同的`AuthMethod`，然而在这一阶段，服务端还不知道要连接的用户是哪一个，自然不知道正确的`AuthMethod`应该是什么了）

客户端根据`AuthMethod`定义的方法对密码+payload加以计算，计算结果连同`AuthMethod`一起放在`HandshakeResponse`里一起发给服务端。

如果服务端读取对应的用户表之后，发现`AuthMethod`跟猜测的一致，那么就可以直接验证客户端的计算结果了，成功后直接返回`Ok`，这样就完成连接建立了。否则，服务端需要发送`AuthMethodSwitchRequest`来重新进行鉴权。

## tidb-gateway的实现

Gateway的实现基本上就是经典的man-in-the-middle，在客户端和后端TiDB之间相互转发消息，顺便在中间做一些手脚。不过，首先需要解决的问题是，怎么获得连接对应的是哪个用户集群来进行路由。

### 传递cluster id

对客户端来说，它仍然是以连接MySQL Server的方式在连接Gateway，所以我们需要想办法在协议中插入集群信息。

MySQL的`HandshakeResponse`中有个`Attrs`字段可以用来插入一些自定义信息，可惜不是所有的DB Driver都支持设置。权衡之下，我们最后决定直接把集群id跟用户id拼接在一起，比如默认的root用户改成`{clusterid}.root`，这样虽然看上去有点怪，但是能保证兼容所有的客户端。

### 连接建立过程

这个过程比较显然了：

1. 客户端向Gateway建立TCP连接。

2. Gateway构造一个默认的`InitialHandshake`消息返回给客户端。

3. 客户端发送`HandshakeResponse`消息给Gateway。

4. Gateway解开`HandshakeResponse`，如果设置了`ClientSSL`此处将连接升级成TLS连接。

5. Gateway根据`UserName`设置的clusterid找到用户集群发起TCP连接，此处如果集群处于休眠状态要先唤醒。

6. TiDB向Gateway发送`InitialHandshake`。

7. Gateway把从客户端收到的`HandshakeResponse`发送给TiDB。

8. Gateway把两个连接串连起来对拷数据。

### AuthMethod的特殊处理

由于MySQL协议中鉴权FastPath的存在，这个过程是有问题的：客户端收到的`challenge payload`是一开始由Gateway生成的，它跟后端TiDB发给Gateway的显然不一致，这将导致后端TiDB在收到`HandshakeResponse`后校验失败报错。

不过，校验失败的前提条件是FastPath被成功激活，即TiDB初始猜测的`AuthMethod`是正确的，否则TiDB不会激活FastPath，而是发送`AuthMethodSwitchRequest`尝试重新鉴权。

解决这个问题的方法也很简单，我们把转发给TiDB的`HandshakeResponse`篡改一下，改成一个TiDB不认识的`AuthMehod`，这样FastPath就不会激活了。

### TLS的特殊处理

因为Gatway和TiDB的连接是在足够安全的内网，从节约能源的角度考虑，我们希望避免在Gateway和TiDB使用安全连接。

这样就带来一些问题：在客户端看来，它跟服务器之间是安全连接，但是在TiDB看来，连接是非安全的，会产生一些不一致的现象。比如`require_secure_transport`功能（这个选项限制TiDB只接受安全连接）就不能用了，还有系统表中Ssl相关的信息显示也都不正常。

解决办法是利用了MySQL Protocol的那个可以在插入自定义`Attrs`的功能，由Gateway把客户端连接的TLS相关信息通过`Attrs`发送给TiDB，然后我们给TiDB打了个小补丁，让它可以把TLS信息解析出来，并设置上安全连接的标记。

### 数据压缩和sequence number

MySQL协议支持设置数据压缩，可以在进行导入导出等场景下显著节约流量。与TLS类似的，我们也希望数据压缩只在客户端和Gateway之间启用，Gateway和TiDB之间保持关闭以减少TiDB的CPU消耗。

不过，MySQL Protocol中有一个sequence number的概念，它需要被携带在每个消息包中，并且在一次客户端服务器交互过程中保持+1递增。譬如，客户端向服务器发送一个查询，拆分成2个消息包，sequence number就分别是0、1，服务器返回2次result，拆分成3个消息外，sequence number分别是2、3、4，然后客户端发送下一轮查询，再从0开始重新计数。

当Gateway两端的压缩方式不一致时，拆分包的粒度不一样，会产生sequence number对不上的情况。所以这种情况下，就不能简单地做data stream拷贝了，而是要认认真真把每个消息包解出来，并在两端分别维护sequence number。

这块还是比较麻烦的，尤其是sequnce number的处理，这里就不细说了，感兴趣可以参考下具体代码。

## Gateway开发和上线过程

上面说的这些功能是分了几个版本迭代出来的，大体过程是有些进展了就发布一次，完善和修完bug了再继续做下个版本。

第1版只有基础的代理功能，上线之后替换掉了每个集群的公网LB。

第2版加上了唤醒用户集群的功能，同时大神同事也做了一些神奇的K8s优化，把唤醒用户集群的耗时从几分钟降到了10几秒，于是我们把休眠时间从7天逐渐降到了小时级别。

第3版是去掉了内网流量的TLS。

数据压缩功能目前看还不太需要，所以一直没有开启。

## 下一步开发计划

### 共享数据集

DevTier有一个问题是给的配置太低了，容量也小，所以很难体验到TiDB在大数据下的表现，比如HTAP特性什么的。

我们有个想法，就是在服务器上预先搭一套高配集群，然后提前灌一些数据进去，比如ossinsight用的github\_archive。

用户通过Gateway连接上自己的集群后，Gateway监听客户端发过来的消息包，如果发现用户use特定的database，就把流量转接到共享数据库，这样用户就可以很方便地体现TiDB的一些特性了。

这个功能现在已经在Gateway上差不多实现出来的，不过后面要不要上还不好说。

### serverless支持

serverless和多租户是TiDB Cloud未来的一大演进方向，架构上简单说就是很多用户共享一套TiKV集群，然后为每个用户启动单独的tidb-server，具体的思路可以参考下[天才阿毛的blog](https://www.zenlife.tk/tidb-multi-tenant.md)。

在这套架构下，用户连接上来时只用启动tidb-server就行了，如果容器和进程都提前启动好，tidb-server的初始化过程在数百毫秒内就能完成。

有了这个速度，我们的休眠-唤醒策略就可以做得更激进了，比如由Gateway维护一个TiDB Pod池，当收到用户连接时从池中抓一个Pod出来使用，用户连接一断开就立即退出返还回池子里。

## 广而告之

由于边际成本的逐步降低，我们的DevTier服务不设一年的使用限制了，注册一下就可以拥有一个长期免费的TiDB集群，虽然配置是差点，跑些个人小应用还是很合适的，欢迎来玩！注册地址在这里：[https://tidbcloud.com/](https://tidbcloud.com/)
