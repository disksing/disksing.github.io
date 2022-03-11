---
title: "Go语言泛型初体验"
date: 2022-03-11
tags: ["编程语言","Go"]
---

Go1.18rc1 放出来也有一段时间了，我们期待了多年了泛型的支持终于是要实装了，毕竟已经是RC，后面语法应该不会再大动了，所以决定提前来学习一下。

前几年曾经用Go语言移植了C++ STL的迭代器和算法库（[disksing/iter](https://github.com/disksing/iter)），因为当时没有泛型，所以基本上是 interface{} 和 type assertion 满天飞的状态。这次我就用它来学习泛型，试着改个泛型版本出来。在C++里面，迭代器和算法这块可以说是泛型应用的典中典，所以我觉得要是能把它给改完，应该能说明这一版泛型的实用程度是很高的了。

先说结论吧，我觉得至少可以打85分。中间确实也遇到一些障碍和小的体验问题，但是瑕不掩瑜，我觉得这一版泛型在“保持简洁”和“提供更完善的功能”间保持了非常好的平衡。几乎不需要了解什么额外的概念和实现原理，就凭着自己对泛型朴素的理解，就能比较顺利地上手了。

最简单的基础用法这里就不多说了，有兴趣的话可以参考下官方blog的那篇文章。这里仅挑我遇到的几个问题分享一下。

### 自指

有时候我们需要在 interface 中定义与具体类型相关的方法，比如 Copy() 用于复制一个同类型的对象，或者 Next() 用于返回指向下一个位置的迭代器，又或者 Equal() 用来和同类型的对象进行比较。

在 Rust 里面有一个 `Self` 来解决这个种问题。在 impl 的时候，你的具体类型是啥，就返回啥。

```rust
trait Copyable {
    fn copy(&self) -> Self
}
```

在 Go 里面，没有泛型之前，我们一般是这么干的：

```go
type Copyable interface {
    Copy() Copyable
}
```

不过这不是泛型，只是一个常规的 interface。我们在实现具体 struct 的时候，Copy() 只能返回 Copyable 而不能用具体类型，在使用的时候还需要强转一下。

```go
type myType struct{}

func (t myType) Copy() Copyable {
	return myType{}
}

func main() {
	x := myType{}
	y := x.Copy().(myType)
}
```

如果我们想复制一个slice，只能写个这样的：

```go
func copySlice(s []Copyable) []Copyable {
	s2 := make([]Copyable, len(s))
	for i := range s {
		s2[i] = s[i].Copy()
	}
	return s2
}
```

这个其实基本上没啥用，因为想处理具体类型的时候，得先转成interface的版本，复制完了还得再转回去……

要用上泛型，也就是要引入类型 T 嘛，一般人都会想这么来：

```go
type Copyable[T Copyable] interface {
	Copy() T
}
```

但是报错，说 type Copyable 不能自指：`invalid recursive type Copyable`。

真正的解决方法比较奇妙。我们先做这么一个泛型 interface，同时 myType 就按我们的想法直接返回它自己：

```go
type Copyable[T any] interface {
	Copy() T
}

type myType struct{}

func (t myType) Copy() myType {
	return myType{}
}
```

这里 Copyable 这么定义显然是没问题的，它不涉及到自指，实际上这个 interface 的意义是“Copy()函数返回一个随便什么类型”——具体的类型可以实例化的时候再决定。

那么这里 myType 有没有实现 Copyable 呢？是有的，当 T=myType 的时候，也即 myType 实现了 `Copyable[myType]`。

下面关键的来了，看一下泛型 copySlice 的写法：

```go
func copySlice[T Copyable[T]](s []T) []T {
	s2 := make([]T, len(s))
	for i := range s {
		s2[i] = s[i].Copy()
	}
	return s2
}
```

最妙的是这里的 `T Copyable[T]`，它指明了对类型 T 约束是 T 要实现 Copyable[T]，上面我们已经说过了 myType 实现了 `Copyable[myType]`，因此这个泛型函数确实可以接收 []myType 并返回一个 []myType。

老实说，我感觉这个用法是目前Go泛型最玄妙的地方，我至今没有完全想明白，尤其是在使用场景复杂了以后。比如 iter 项目中的[这个地方](https://github.com/disksing/iter/blob/937b0b8f9ffa1df6d50bb045a89a977c51a97f61/iterator.go#L129-L132)，这里 first 和 last 是一样的类型（都是 RandomIter，但是这里类型转换的时候必须前一个转成 RandomIter，后一个转成 It，不然下面的 `f.Distance(l)` 编译不过。如果有完全想透彻了的欢迎分享一下……

### 运算符重载

在 C++ 里面，我们可以使用运算符重载来让自定义类型支持像数字一样的运算，比如可以给矩阵重载一个加法运算，这样我们可以写一些算法（比如map-reduce），让它可以同时用在基本类型和自定义类型上面。

不过在当前Go语言泛型里还不支持运算符重载，大体上你只能再搞个“假自指”的接口，要求具体类型实现某个方法，比如实现 Less 来代替 "<"：

```go
type Ordered[T any] interface{
	Less(T) bool
}
```

这样确实可行，不过问题是泛型接口不支持定义成“是某些类型**或者**实现了某些方法”，如果你尝试写成这样：

```go
type Ordered[T any] interface {
	~int | ~string
	Less(T) bool
}
```

它的意思实际上是“(类型是int或string)**并且**实现了Less方法”，我们不可能用它来实现同时作用于基本类型和自定义类型的泛型代码。

这个问题目前来看应该是无解的，我们不得不定义两个函数来分别供基本类型和自定义类型使用。

不过我们未必需要把同样的代码写两遍，一个相对优雅一点的方法是，把运算符操作当成一个额外的参数传出泛型函数，实际上C++ STL里很多地方都是这么干的。

比如我们想实现一个`Min`函数返回两个变量的较小值，可以写成这样：（在 iter 项目中，我总是使用 By 后缀来表示支持传入自定义运算符操作）

```go
func MinBy[T any](a, b T, less func(T, T) bool) T {
	if less(a, b) {
		return a
	}
	return b
}
```

针对基本类型的 `Min` 函数也不用把 `MinBy` 的函数体抄一遍，它可以调用 `MinBy`：（因为 Min 的 T 是 Ordered，范围小于 MinBy 的 any）

```go
func Less[T Ordered](a, b T) bool {
	return a < b
}

func Min[T Ordered](a, b T) T {
	return MinBy(a, b, Less[T])
}
```

另外，使用自定义类型时，类型的方法也是能当作最后那个操作符函数使用的：

```go
type myType struct {
	x int
}

// 不是一定要 func Less(t1, t2 myType) bool
func (t myType) Less(t2 myType) bool {
	return t.x < t2.x
}

x := myType{x: 1}
y := myType{x: 2}
MinBy(x, y, myType.Less)
```

### 特化

特化是C++模版里面的一种高级特性，大致就是在泛型代码中，可以针对具体实例化的类型写一些特殊的逻辑。

它常常用来做性能优化，比如STL中的[Sample算法](https://en.cppreference.com/w/cpp/algorithm/sample)，如果输入不是 RandomReader（说明计算总长度成本高）且输出支持随机写，那么会使用蓄水池算法，否则会先算总样本数，再按概率直接选。

不用看文档我们就知道，Go语言肯定不可能有这个……不过在前泛型时代，我们一直有在用的是接口的 type assertion，包括两种形式，`t, ok := iface.()` 和 `switch iface.(type) {}`。

在泛型代码中，实例化过后，变量都是具体类型了，没有interface。不过没有条件创造条件也能上，奇技淫巧来了，我们只需要把具体类型转成interface{}，然后就可以愉快地判断类型了。更妙的是，因为Go1.18引入了any关键字，我们甚至不用写interface{}……

这个比较简单就不多说了，贴一段 iter 里的代码吧。我觉得这个尽量少用，因为一旦用上了这招，各种前泛型时代的强转就全来了。

```go
// AdvanceN moves an iterator by step N.
func AdvanceN[T any, It Iter[T]](it It, n int) It {
	if it2, ok := any(it).(RandomIter[T, It]); ok {
		return it2.AdvanceN(n)
	}
	if it2, ok := any(it).(ForwardIter[T, It]); ok && n >= 0 {
		for ; n > 0; n-- {
			it2 = (any)(it2.Next()).(ForwardIter[T, It])
		}
		return it2.(It)
	}
	if it2, ok := any(it).(InputIter[T, It]); ok && n >= 0 {
		for ; n > 0; n-- {
			it2 = any(it2.Next()).(InputIter[T, It])
		}
		return it2.(It)
	}
	if it2, ok := any(it).(BidiIter[T, It]); ok && n <= 0 {
		for ; n < 0; n++ {
			it2 = any(it2.Prev()).(BidiIter[T, It])
		}
		return it2.(It)
	}
	panic("cannot advance")
}
```

### 总结

再接再厉，还需要再加强理解……