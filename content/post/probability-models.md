---
title: "二项分布，泊松分布和指数分布"
date: "2020-02-05"
tags: ["数学", "概率论"]
draft: true
---

为了研究这个问题，我花了不少精力复习差不多忘光的概率论，这里先梳理一下相关知识点。

回忆当年《概率与统计》课程，依稀记得说电灯泡的损坏是服从[指数分布](https://zh.wikipedia.org/zh-hans/%E6%8C%87%E6%95%B0%E5%88%86%E5%B8%83)的，网上查了下分布函数是这样的，表示灯泡寿命小于 x 的概率：

$$
F(x) = 1 - e^{-\lambda x}
$$

不得不说看到这个我感觉一头雾水，这个公式是怎么来的呢？公式里的$\lambda$表示啥？$e$又是从何而来？又是一顿查资料后，发现还是比较好搞理解的。

灯泡寿命的模型也是简化过的，模型认为灯泡的损坏是没有记忆性的，也就是每时每刻灯泡故障的概率是一样的，跟灯泡用了多久没关系。这就相当于每隔一小段时间掷个骰子判断灯泡会不会损坏，实际上就变成了[二项分布](https://zh.wikipedia.org/zh-hans/%E4%BA%8C%E9%A0%85%E5%88%86%E4%BD%88)，如果进行 N 次随机独立实验，每次成功的概率为 p，总共成功 k 次的概率按如下公式计算：

$$
P(X=k)=\binom{n}{k}p^k(1-p)^{n-k}
$$

假设单位时间的灯泡的损坏率是$\lambda$，平均分成 n 份之后，每一小段时间的损坏率就是$p=\frac{\lambda}{n}$，当 n 趋近于$+\infty$的时候，把上面的公式化简：

$$
\begin{eqnarray}
\lim_{n \to \infty}P(X=k)
&=& \lim_{n \to \infty}\binom{n}{k}p^k(1-p)^{n-k} \\\\
&=& \lim_{n \to \infty}\frac{n!}{(n-k)!k!}(\frac{\lambda}{n})^k(1-\frac{\lambda}{n})^{n-k} \\\\
&=& \lim_{n \to \infty} \underbrace{\frac{n!}{n^k(n-k)!}}_{1} (\frac{\lambda^k}{k!}) \underbrace{ (1-\frac{\lambda}{n} )^n}_{e^{-\lambda}} \underbrace{ (1-\frac{\lambda}{n})^{-k}}_{1} \\\\
&=& \frac{\lambda^k e^{-\lambda} }{k!}
\end{eqnarray}
$$

这其实就是[泊松分布](https://zh.wikipedia.org/zh-hans/%E6%B3%8A%E6%9D%BE%E5%88%86%E4%BD%88)在 t 取单位时间时的特殊形式，泊松分布用来描述随机事件在一定时间内发生指定次数的概率，常常用来估算如公交车到达，小孩出生，设备损坏等问题，其一般形式是这样的：

$$
P(N(t)=n)=\frac{(\lambda t)^ne^{-\lambda t}}{n!}
$$

其中$\lambda$是泊松过程的强度，也就是单位时间内事件发生的期望次数。接下来我们应用泊松分布来求随机事件在时间 T 内发生至少一次的概率：

$$
\begin{eqnarray}
F(t) &=& P(N(t)\ge 1) \\\\ &=& 1-P(N(t)=0) \\\\ &=& 1-\frac{ (\lambda t)^0 e^{-\lambda t}}{0!} \\\\ &=& 1-e^{-\lambda t}
\end{eqnarray}
$$

这样我们就得到了最开始看到的那个指数分布公式。之前的几个问题也有了答案，$\lambda$是单位时间的灯泡损坏率，而自然底数$e$则是在从离散到连续的“无穷细分”过程中引入的。
