---
title: "beancount 复式记账实践"
date: 2020-08-20
tags: ["杂谈"]
---

尝试复式记账（beancount）有一段时间了，也有了一些实践上的体会（和走弯路），本文就简单做个分享。注意，其实我的很多做法完全算不上专业，甚至十分粗糙，不过我感觉自己简单用用还是可以的，操作起来也比较简单，主要供刚入门的小白参考一下吧。

## 入门

入门向的中文资料和种草文章有很多了，我这里不准备献丑，扔几个链接算了：

- [Beancount —— 命令行复式簿记 by wzyboy](https://wzyboy.im/post/1063.html)
- [beancount 起步 by MoreFreeze](http://morefreeze.github.io/2016/10/beancount-thinking.html)
- [Beancount复式记账 by BYVoid](https://byvoid.com/zhs/blog/beancount-bookkeeping-1/)
- [复式借贷记账法 by ShanHe Yi](https://yishanhe.net/beancount-tutorial-0/)

有两个话题大家几乎都有谈到，分别是为什么要用复式记账，以及为什么要用beancount，我也简单谈下感想好了。

在我看来，相对于普通的单式记账（或叫流水账），复式记账的主要优势在于更能体现资金流动的本质。

假如我花15万买了一辆车，开了3年以后卖掉得10万。如果记流水账，从账面上的钱来看，3年前买车花了15万，3年后卖车赚了10万。这似乎符合我们日常生活的认知，但没触及本质。如果使用复式记账，你会得到另外一个版本的故事：3年前钱没有花掉，只是把15万现金换成了固定资产——汽车，而3年后也没有突然获得不菲收入，只是把汽车又变成了现金，这二者差的5万，是在3年不断使用过程中慢慢花掉的。

基于更符合本质的账目，我们将有机会深刻洞察自身财务状况，做出明智的决定。比如上面那个买车卖车的例子，我们会意识到机动车的保值率可能比其绝对价格更能影响其实际产生的花费。另外一个例子是，当我把每个月还的房贷拆解成本金和利息之后，从报表上意外地发现房贷利息占了每月开支的相当一部分，于是促成了我把提前还一部分房贷提上日程。

至于说复式记账软件有很多，为什么要用beancount。我认为beancount主要是对有编程经验的人比较友好，因为是纯文本的，可以很方便地做各种转换，不仅可以自己写一些脚本来自动生成账目，也能把beancount账目导出到别的系统。如果不懂编程的话，就不是很推荐了，图形化界面的软件可能更合适。

## 工作流

我的做法其实是比较山寨的，很多高级功能都没用。比如我只有人民币一种货币，其他的货币全都转成人民币的入账，再比如我的所有账目都记录在一个单一的beancount文件里（为了方便导出到其他系统）。

我是每个月找一天来集中记账的，平时就完全不考虑记账的事情（感谢无现金时代）。我的情况是每月底至下月初资金变动会比较大，包括发工资、信用卡还款、还房贷，等这些“尘埃落定”后，我就会找比较空闲的一天来记账，记完之后顺便就把资金归置归置，比如不急用的钱扔到余额宝。

第一步是处理微信账单。我的绝大多数交易都是使用微信支付的，包括信用卡也是通过微信来付，主要是因为微信的账单功能特别好用，导出的账单是一张尤为详细的excel表，可以很方便地进行处理。我写了一个简单的脚本，能把excel转换成beancount格式，而且能识别常用的收款方。识别不了的，就需要手动过一遍，标上正确的花费类型。这里常常会遇到想不起来花的钱是怎么花的的情况，可能需要去京东上查订单，或者查当天的聊天记录，或者查当天的日记，一般情况下都是能想起来的。

第二步是所有的银行卡。因为基本上都走微信了，剩下的一般包括工资、房贷、转账，还有少量支付宝的花费和少量的存款利息收入。这里基本就是打开手机网上银行，然后手动录入。这里隆重推荐一下云闪付APP，绑定银行卡之后，一个页面就能显示所有卡的余额了，对账十分方便。

第三步是支付宝。可能会少量付款是用的支付宝付的，需要手动登记一下，还有就是余额宝的利息收入了。

第四步是在公司吃饭的园区卡。这一步是我现在最痛苦的了，消费记录可以在一个APP上查到，但是可惜不能导出，我尝试用fiddler抓包不过也可耻地失败了。现在我的权宜之计是用手机打开消费记录页面，滚动截屏，OCR，再拷贝到电脑上手动调整下格式和识别错误的内容，再用脚本处理一下。

第五步就是各种充值卡了，包括京东E卡，kindle余额，steam余额之类的。这些变动比较少，我处理的比较随意了，我一般有记得的的就打开对应的订单记录一下，记不清就算了，等下次发现对不上的时候再补上。

## 记法实践

最后分享一下我摸索的一些常见事项的记法吧，仅供参考。

### 工资收入

工资其实可以记的很细的，比如把五险一金的详细情况都记下来。我对交了多少社保多少个税没太多执念，所以就直接记税后工资了。不过我工资卡、公积卡、医保卡分别是不同的卡，所以我把工资也简单地分成这三块了。

### 房贷

房贷上面也提过了，记账的时候注意要把交的钱拆成本金和利息，如果直接全还到负债里面，是平不了账的。利息部分我是记成花费（Expenses）的，貌似也有别的记法，不过我觉得记成花费没什么毛病，不深究了。

### 报销

报销跟前面那个买车卖车的例子有些类似。看似是我花了钱，但是因为这个钱后面公司是给报销的，所以记成花费就不太合适了。我的做法是在资产里用公司的名字建一项应收款账户，花钱的时候，不记花费而是记入应收款，到时候公司给报销了，再把钱从应收款里转出来就行了。同时，检查应收款就能很方便地知道还有多少钱没报销。

### 信用卡

信用卡一度让我相当凌乱，主要是信用卡本身就有对账日和还款日，然后我每个月的对账日还跟这两个都不一样，在对账日当天很难搞明白我到底应该欠银行多少钱。

后来我重新整理了思路，意识到账目的建模应该跟信用卡本身的内在逻辑是匹配的，信用卡其实同时是存在两个账户的：一个是上月账单一个是本月账单。因此我照葫芦画瓢，给每张信用卡建个应付款子账户。同时，记账当天不对信用卡进行对账了，而是在信用卡的对账日来对账，具体过程是先检查负债是否能对上银行的账单，确认后把所有的钱从信用卡主账户转进应付款，而每当记到还款日时，则清空应付款。

