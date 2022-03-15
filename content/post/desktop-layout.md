---
title: "适合程序员的桌面窗口管理方案"
date: 2022-03-15
tags: ["教程"]
---

介绍下我在办公室使用的桌面窗口管理方案。

先说下硬件：我的工作电脑是一台 13 英寸的 Macbook Pro，外接了一个 27 英寸的显示器，使用了一个电脑支架把笔记本架起来了，这样两个显示器差不多是并排的样子。

然后看下我的窗口布局规划，可以会比较特殊一点，不过是经过精心考虑的。

{{< figure src="/assets/img/desktop-layout.jpg" title="桌面布局示意" style="text-align:center" >}}

右边的大显示器我按 1:2 分成两列，左边这一列是最大的显示空间，我把它当作“主空间”来使用，一般我在编码状态下会把编辑器放在这个位置，非编码状态下这里通常就是浏览器了。因为本来两个显示器的大小是不一样的，这样进行划分，主空间两侧的空间反而是比较均衡的状态了。显示器和电脑屏幕在桌子上的摆放也是非对称的，当我坐下时，正对着的是这个“主空间”的正中央，这样也避免一个常见问题：两个一样大的显示器对称排布时，正对着的位置恰好是两个显示器中间的缝，于是工作中几乎时刻是扭着头的，时间一长脖子就受不了了。

左边笔记本的屏幕没有做切分，这块空间一般在编码的时候放浏览器看文档资料，或者放个 Terminal 调试，也可以是另一份代码，使用完整的屏幕保证它总是够用的，不至于不得不把窗口拉大然后频繁切换窗口。

大显示器右边的 1/3 被一分为二，这两块空间主要用来放IM软件，包括飞书、Telegram、微信、QQ、Twitter桌面版……IM软件也平铺出来也是为了减少切换窗口，我只需要在干正事的时候时不时瞟一眼就行了，有需要关注的消息时再去处理。必要的时候这个小格也可以临时放一下 Terminal 之类的小窗口。

再说窗口管理软件方面，可能我的搞法比较变态，我也没找到合适的软件，最后使用的方案是 HammerSpoon 一点脚本，HammerSpoon 大体上就是 Mac 版的 AHK，功能是弱了很多，不过在窗口管理这一块还是完全够用的。我的脚本很简单，使用 4 个快捷键，分别把窗口移动到 4 个格子：

```lua
hs.hotkey.bind("cmd", "1", function()
    local sf = hs.screen.primaryScreen():frame()
    hs.window.focusedWindow():setFrame(hs.geometry.new(sf.x, sf.y, sf.w*2/3, sf.h))
end)

hs.hotkey.bind("cmd", "2", function()
    hs.window.focusedWindow():setFrame(hs.screen.allScreens()[2]:frame())
end)

hs.hotkey.bind("cmd", "3", function()
    local sf = hs.screen.primaryScreen():frame()
    hs.window.focusedWindow():setFrame(hs.geometry.new(sf.x+sf.w*2/3, sf.y, sf.w/3, sf.h/2))
end)

hs.hotkey.bind("cmd", "4", function()
    local sf = hs.screen.primaryScreen():frame()
    hs.window.focusedWindow():setFrame(hs.geometry.new(sf.x+sf.w*2/3, sf.y+sf.h/2, sf.w/3, sf.h/2))
end)
```

最后还有一个问题，就是我平时需要启动的 IM 太多了，虽然大部分消息都是不需要实时处理的，但是为了不错过重要消息，不得不隔一段时间手动检查一遍。解决这个问题我还是用 HammerSpoon 脚本，我会把各种不需要实时关心的IM窗口全扔到4号格子里，然后用脚本开个定时器，隔一段时间就把这一坨里在最下面的窗口给捞到最前面。3号窗口是固定不切换的，预留给比较重要的IM窗口，比如飞书。

这个切换窗口的 HammerSpoon 脚本也分享一下：

```lua
hs.timer.doEvery(120, function()
    local sf = hs.screen.primaryScreen():frame()
    local windows = hs.window.orderedWindows()
    local bottom_win = nil
    for i, win in ipairs(windows) do
        local fm = win:frame()
        if math.abs(fm.x-(sf.x+sf.w*2/3))<1 and math.abs(fm.y-(sf.y+sf.h/2))<1 then
            if win:id() == hs.window.focusedWindow():id() then
                break
            end
            bottom_win = win
        end
    end
    if bottom_win ~= nil then bottom_win:raise() end
end)
```
