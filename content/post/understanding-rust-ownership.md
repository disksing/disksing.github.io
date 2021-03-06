---
title: "五句话理解 Rust 所有权"
date: "2020-02-08"
tags: ["编程语言", "rust"]
---

_先免责声明一下哈，我是 Rust 新手入门，也没研究过 Rust 编译器。本文只是我自己学习 Rust 所有权时的一些思路，或者说对相关概念的一种解释吧，仅供参考，有谬误在所难免，还请指正。_

### 1. 所有权检查在编译期约束变量名如何访问资源。

所有权检查是编译期的静态检查。这意味着它将不会带来任何运行时的开销。但同时需注意编译器通常不会考虑你的程序将怎样运行，而是基于代码结构做出判断，这使得它经常看上去不那么聪明。

比如你依次写两个条件互斥的 if，编译器可不想那么多，直接告诉你不能 move x 两次。

```rust
fn foobar(n: isize, x: Box<i32>) {
    if n > 1 {
        let y = x;
    }
    if n < 1 {
        let z = x; // error[E0382]: use of moved value: `x`
    }
}
```

甚至你把 move 操作放在循环次数固定为 1 的 for 循环里面，编译器也傻傻看不出来：

```rust
fn foobar(x: Box<i32>) {
    for _ in 0..1 {
        let y = x; // error[E0382]: use of moved value: `x`
    }
}
```

我们常用“变量”这个术语，比如`let x = Box::new(1i32)`，被称之为“定义值为 1 的变量 x”，这是比较简化和高度抽象的说法，实际情况更接近于“告知编译器，在此处需要在堆上分配长度为 4 字节的内存并按照 32 位 int 初始化为 1，此块源代码我将使用 x 来指代这块内存”。

在这个例子里，那一块在运行时将被分配出来的内存就是资源，而字符串 x 就是变量名。你可以想像编译器在运行过程中，对当前 scope 可见的变量名维护着一个状态表，并针对每行源代码检查这个变量名的使用是否符合它当前的状态。

编译期的约束是作用在变量名上的，而不是那块内存。比如上面的例子中 x 是只读的，并不妨碍我们定义另外一个可写的变量名来写这块内存（`let mut y = x`），甚至重新定义 x 来为可写的变量名（`let mut x = x`）。

### 2. 资源有且仅有一个 Owner，资源在 Owner 作用域结束时资源被释放。

显然，这个规则是为了防住 C/C++ 中 new 和 delete 不匹配造成的空悬指针、多次释放、内存泄漏等问题。Rust 的解决思路是把资源的生命周期跟变量名的作用域强行绑定在一起，并强行规定只有一个 Owner。很容易证明：1）只有一个 Owner，资源不会被释放多次；2）只有一个 Owner，资源释放后不会被其他的变量名与之绑定。大体来说，就按照 C++ 的 unique_ptr 来理解就对了。

Owner 是可以转移的。`let y = x` 将 x 所绑定的资源转移给 y，之后 y 就变成了新的 Owner，而 x 的状态变成了 moved，无法再进行读写。对比 C++ 的话，`let y = x` 相当于 `auto const y = move(x)`。

### 3. 资源可以有多个引用，引用变量名的作用域不能超出 Owner 的作用域。

引用我们一般在函数调用的时候用，它不太容易从语言层面省掉，因为如果没有引用，在调用函数的时候就只能选择 move 进行函数，然后再用函数返回值的方式“返还”给上层，否则资源会随着函数的结束而被销毁。

Rust 里引用被叫做借用（Borrow），也暗示着所有权被（部分）转移随后返还的过程，返还发生在引用变量名作用域结束的时候。

引用变量名的作用域不能超过 Owner 的作用域，这个几乎就是理所当然的了。如果没有这个限制，将可能访问到已经被释放的资源，这也是 C++ 中另一种常见的空悬指针的产生原因。

### 4. 同一时刻，一份资源只能被至多一个变量名读写，或者被多个变量名读取。

这个规则比较有意思。看上去有些像读写锁，但是我们知道读写锁是为了防止并发场景下多线程访问同一个变量产生的冲突问题，那在非多线程场景下这个规则有什么用呢？

还是以一段 C++ 举例：

```cpp
if (a == 0) {
    b++;
    cout << a << endl;
}
```

乍看上去，这段代码要么没输出，要么输出 0，其实不然，比如 b 恰好是 a 的引用……在现实场景中，`b++`这一行可能是一个错综复杂的函数调用，如果出现对 a 的非预期修改，那就更难排查了，这正是很多函数式编程所倡导的 immutability 的好处。Rust 在规则的限制下，也规避了这个问题：此处 a 是可读的，所以 a 所绑定的资源不可能被其他变量名修改。

我们根据这个原则，很容易就能排出来借用时的具体规则：

对于不可变借用（`let y = &x`）：1）引用 y 是只读的；2）在 y 作用域结束之前，x 可读不可写（因为存在 y 可读），x 能被不可变借用（因为没有变量名能写）但不能被可变借用（因为存在 y 可读）。

对于可变借用（`let y = &mut x`）：1）引用 y 可读写；2）在 y 作用域结束之前，x 不可读不可写（因为存在 y 可写），x 不能再次被借用（因为存在 y 可写）。

### 5. Rc 用于解除“只有一个 Owner”的限制，Cell 用于解除“&T 不可写”的限制。Arc 是线程安全的 Rc，RwLock/Mutex 是线程安全的 Cell。

到此为止，这一套机制看上去很美好，可惜理想很丰满现实很骨感。只要写一写稍具规模的代码，就会不可避免地发现：我需要突破编译器的限制！比如：“我的多个引用都需要时不时写一下，但是我的逻辑能保证不会几个地方同时写”，或者“我就是需要有多个可写的引用，我的逻辑能保证不了问题”，又或者“我的资源是多个 struct 间共享的，没有明确的 Owner，特别是我不知道这几个 struct 哪个先销毁”。（注意这些例子都不涉及到多线程）

这些情况下，Rust 编译器的态度是一贯的：我感觉你想干坏事但是没证据，先报个错再说。此时 Rust 的 NB  之处就体现出来了，不仅提供 unsafe 用于绕过编译器的种种限制，还在标准库中提供了一堆设施让你更方便地以各种姿势绕后……

Rc 用于解除“只有一个 Owner” 的限制。资源将在最后一个 shared owner 作用域结束时被释放，不难看出实际上就是 C++ 的 shared_ptr。老实说，这了所谓的安全和优雅在语言层面只支持 unique_ptr，然后不得不在标准库提供 shared_ptr 来绕过语言的限制，不见得有多高明。

几种 Cell（Cell，RefCell，UnsafeCell）用于解除“&T 不可写的限制”，实现上很简单，用 unsafe 把 &T 转成 &mut T。换个角度来看，也可以理解成解除了“&mut T 只能有一个的限制”。其中 Cell 和 RefCell 把编译期的检查移到了运行时，也就是说仍然会在运行时检查条款 4，如果被违反了会产生 panic。而 UnsafeCell 就比较粗放了，可以随便转，用户自己保证逻辑安全。

Arc 是线程安全的 Rc 这个大家都知道，不细说了。Mutex 和 Cell 的相似性似乎还没人提出来过，我觉得它们的内在逻辑是非常类似的：多个变量持有资源的不可变引用，在保证安全的情况下，把不可变引用转成可变引用并进行读写。区别仅在于多个变量是在同一线程还是不同线程，以及由用户的逻辑保证安全，还是由操作系统 API 保证安全。

---------

最后让我们再复习一下吧：

1. 所有权检查在编译期约束变量名如何访问资源。
2. 资源有且仅有一个 Owner，资源在 Owner 作用域结束时资源被释放。
3. 资源可以有多个引用，引用变量名的作用域不能超出 Owner 的作用域。
4. 同一时刻，一份资源只能被至多一个变量名读写，或者被多个变量名读取。
5. Rc 用于解除“只有一个 Owner”的限制，Cell 用于解除“&T 不可写”的限制。Arc 是线程安全的 Rc，RwLock/Mutex 是线程安全的 Cell。
