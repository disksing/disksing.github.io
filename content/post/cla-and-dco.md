---
title: "当我们签 CLA 或 DCO 的时候，我们在签什么"
date: 2020-01-28
tags: ["开源"]
---

有些开源项目会要求 contributor 在提交 PR 的时候签署 CLA 或者 DCO，很多人（比如说我）按照提示进行一些操作也就签了，但是类似于安装软件的时候点“我已阅读并同意”，实际上并不知道自己签了啥。趁着放假加上新型肺炎的疫情遭遇封城，我研究了下这个问题。

要说清这两个概念，我们先要聊一下开源许可证。这个我们平时接触的很多了，只是有一点尤其值得注意，就是原理上软件开源开放的是**使用权**，而**著作权（又称版权）并没有被放弃**，开源许可证指明了用户在某些限制下使用软件及其源代码，如果用户违反了开源许可证，最终依然要回归到著作权的法律框架下解决争端。

著作权是怎么来的呢？是作品完成时自然产生的，归作品的作者所有。很显然，对于立足于开放协作的众多开源项目来说，著作权并不是单一的人或实体，而是所有的 contributor 都有一份。可以想象，如果发生开源软件被侵权，著作权的分散将对开源软件项目所有方带的维权带来一些麻烦。另外，假如项目所有者想要更换或调整开源许可证，会因为并不持有全部的著作权受到阻碍，还可能由于著作权的原因与 contributor 产生潜在的争议。

CLA 和 DCO 就是用来解决上面说的这些问题的。

先说 CLA，全称是 [Contributor License Agreement](https://en.wikipedia.org/wiki/Contributor_License_Agreement)，简单来说就是项目在接收 contributor 贡献的补丁之前，需要 contributor 签署的一份协议。与开源许可证不同，CLA 是没有标准化的，也就是说每个项目的 CLA 都可能有自己独特的条款。有些项目的 CLA 会要求 contributor 直接声明将著作权转移给项目方，也有些较为宽松，允许 contributor 持有著作权，但是需要授予项目方发布展示复制等权利（[Apache 软件基金会的 CLA](https://www.apache.org/licenses/cla-corporate.txt)是一个例子）。

有批评者警告开发者[永远不要签 CLA](https://drewdevault.com/2018/10/05/Dont-sign-a-CLA.html)，特别是不要放弃自己劳动成果的著作权。一旦通过 CLA 将著作权转移给项目方，项目方就有权在未来更换其他更严格的许可证，甚至直接将项目变成完全闭源，显然这与开源世界自由开放的初衷是背道而驰的。

然后再说 DCO，全称 [Developer Certificate Of Origin](https://elinux.org/Developer_Certificate_Of_Origin)，最初是在 Linux kernel 项目中被引入的，它的条款很简单，其实就是为了“免责”的目的而存在的，也就是让补丁的提交者确认提交的内容是自己创作的，并且了解项目方会如何使用这个补丁。

与 CLA 相比，CLA 几个受到诟病的点在 DCO 这里是不存在的，作为自由软件的中坚力量，Linux 肯定搞不出转移著作权这种操作，而 Linux 使用的 GPL 许可证也限制了无法 relicense。DCO 的签署方式也有所差别，contributor 需要在每个 commit 添加一行 `Signed-off-by: John Doe <john.doe@hisdomain.com>` 来进行确认。

后来，Linus 因为 kernel 所用的版本控制软件用不下去了，用 10 天时间开发出了如今大放异彩的 git，DCO 也被很好的集成在了 git 里：只需要要 git commit 的时候添加 `-s` 选项，`Signed-off-by` 就能很简单地添加在 commit log 里了。

再后来，由于 git 的使用日益广泛，DCO 也更新到 1.1 版本，添加了关于 relicense 的条款。所以目前 DCO 也适用于其他的开源许可证了，我们可以认为它是标准化的、宽松版本的 CLA。标准化意味着 contributor 不用在参与每个项目时都仔细研究一遍条款，宽松意味着 contributor 不用担心自己的劳动成果被利用。一言以蔽之，DCO 基本上是可以无脑签的。

正是由于给 contributor 带来的心智负担小，集成在 git 里面操作方便，越来越多的开源项目选择从 CLA 迁移成使用 DCO，比如 [CHEF](https://blog.chef.io/introducing-developer-certificate-of-origin/) 和 [GitLab](https://about.gitlab.com/blog/2017/11/01/gitlab-switches-to-dco-license/)。

CLA 和 DCO 哪个更好呢？简单说下我自己的看法吧，CLA 由于存在“作恶”的可能性，风评明显不如安全的 DCO 好，看上去 DCO 是未来的趋势。但是宽松自由的 DCO 也并非没有其弊端，比如软件被侵权之后进行维权，肯定不如集中著作权的方式好操作。而且 DCO 每个 commit 都要进行签名，也不能说不麻烦，而 CLA 只是一次性操作就完事了。说了这几点也别认为我是 DCO 黑了，我想如果在我尝试给一个开源软件做贡献的时候，我应该还是期望其使用 DCO 的，哈哈。

最后说一下我查资料过程中注意到的几个细节：

1. 我们的 TiKV 项目现在操作上是同时要求 contributor 签 CLA 和 DCO 的，但是据我观察 DCO 的条款似乎已经包含在 [CLA](https://cla-assistant.io/tikv/tikv) 里了，所以 DCO 应该可以省掉。（也不知道对不对）
2. DCO 的核心是“原创性确认”，这么来看如果我发现同事的 PR 由于没有 sign 而不能合，我是不能擅自帮他加上 `Signed-off-by`的，因为我不知道他提交的东西到底是不是他原创。与之相对的另一种情况，假如有个 contributor 提交了 PR，但是他不知道怎么去 sign commit，那么我可以通过其他途径与他沟通（比如 email 或者 GitHub issue  comment），只要得到他书面形式的确认，我就可以帮他进行 sign 操作了。
3. CHEF 有一个叫 [Obvious Fix 规则](https://docs.chef.io/community_contributions.html#the-obvious-fix-rule)的优化用于降低 contributor “散户”的进入成本。大体就是说如果是 typo fix 等小修改（达不到“创作”的级别，也就没著作权什么事了），可以绕过 DCO 约束。这个想法不错，值得借鉴。

## 参考资料

- [CLA vs. DCO: What's the difference?](https://opensource.com/article/18/3/cla-vs-dco-whats-difference)
- [Don't sign a CLA](https://drewdevault.com/2018/10/05/Dont-sign-a-CLA.html)
- [wikipedia: Contributor License Agreement](https://en.wikipedia.org/wiki/Contributor_License_Agreement)
- [wikipedia: Git](https://zh.wikipedia.org/wiki/Git)
- [许可证兼容性和再次授权](https://www.gnu.org/licenses/license-compatibility.html)

