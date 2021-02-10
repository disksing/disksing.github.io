---
title: "etcd 的 cluster id"
date: 2020-07-27
tags: ["etcd"]
---

找到生成 Cluster ID 的代码，其实不是某种随机法生成的，而是用所有 MemberID hash 出来的。

```go
func (c *RaftCluster) genID() {
    mIDs := c.MemberIDs()
    b := make([]byte, 8*len(mIDs))
    for i, id := range mIDs {
        binary.BigEndian.PutUint64(b[8*i:], uint64(id))
    }
    hash := sha1.Sum(b)
    c.cid = types.ID(binary.BigEndian.Uint64(hash[:8]))
}
```

那么 Member ID 又是怎么来的呢？

```go
var b []byte
sort.Strings(m.PeerURLs)
for _, p := range m.PeerURLs {
    b = append(b, []byte(p)...)
}

b = append(b, []byte(clusterName)...)
if now != nil {
    b = append(b, []byte(fmt.Sprintf("%d", now.Unix()))...)
}

hash := sha1.Sum(b)
m.ID = types.ID(binary.BigEndian.Uint64(hash[:8]))
```

这个保证了 bootstrap 的时候，多个节点生成出来相同的 cluster id。

不过这导致一些问题，比如我用相同的配置多次启动集群，会看到 Cluster ID 不变。更麻烦的情况是，如果启动集群后销毁一半的节点，再用原来的配置新启动一套集群，前后两套集群是可以互相通信的。可以看看[PD 的这个 issue](https://github.com/pingcap/pd/issues/2606)。

解决办法就是设置启动参数 `--initial-cluster-token`，保证每套集群都不一样。