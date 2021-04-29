# 使用 channel 在 Go 语言中模拟迭代器

## 引言

对于很多人而言，Go 语言的一大槽点是缺少类似于 JavasScript 中的 `map`, `filter`, `reduce` “三剑客”这样的集合操作函数。当我们遍历一个序列的时候只能使用 for 循环，显得似乎不那么灵巧。

对此，Go 语言的作者之一，Rob Pike 大神早有回应。他在 [GitHub](https://github.com/robpike/filter) 上公开了一个 Go 语言版的 `map`、`filter`、`reduce` API。并且在说明中写道：
> “我就想看看在 Go 语言中实现这种东西能有多难。并不难。我几年前写了这个，从未发现它哪里有用。我只用 for 循环。而且我建议你也这样做。”

我赞成 Rob Pike 的观点， 毕竟 Go 语言的特色就是大巧不工。使用 `.map` 和 `for` 循环并没有本质上的区别，只是局部代码风格的差异。一律使用 `for` 循环，反而能让我们把精力集中在程序要解决的问题上，而不是在细枝末节上浪费时间。

但是，有两种情形，是仅仅使用 `for` 循环没有办法解决的，Rob Pike 并没有提及它们。那就是进行延迟计算和表示无穷集合。

为了应对这两种情形，很多编程语言都有“迭代器”这一概念。比如 JavaScript 就有[“迭代器协议”](https://developer.mozilla.org/zh-CN/docs/Web/JavaScript/Reference/Iteration_protocols#iterator)。实际上，迭代器也是23个设计模式中的一种（限于篇幅不展开解释了，不熟悉的朋友请自行了解）。迭代器可以抽象地表示一个无穷集合；通过在迭代器上调用 `.map` 等方法，可以实现延迟计算。由于这个模式很常见，不止 JavaScript，很多其他语言都提供了这个“协议”（或者叫接口），比如 C# 中的 `IEnumerable` 接口和 Python 中的 `__iter__/__next__` 魔术方法。

而 Go 语言中是没有原生的迭代器接口的。乍一看，这似乎是 Go 语言一个很大的不足，毕竟迭代器既强大又常用。但是其实仔细想想，Go 语言其实已经悄悄地内置了一个“迭代器”了，只是不那么明显。那就是 channel 类型。channel 和其他语言中的迭代器具有很多相似之处：它们都可以看作一个抽象、有序的元素集合，都有用于遍历的特殊语法（`range`），都有方法表示迭代已经到达尾部（`close()`）。实际上，channel 就是一个伪装的迭代器。

在这篇文章里，我们将使用 channel 和 goroutine 来实现一个迭代器数据类型 `Iter`。希望可以借此帮助大家更加深入的理解 channel 和迭代器之间的相似之处。完整代码请见 GitHub：https://github.com/CoderYihaoWang/goiter

## 实现 `Iter` 数据类型

如前文所讲，channel 就是一个伪装的迭代器。为了突显这一点，我们这样声明 `Iter`：
```go
// iter.go
package main

type Iter <-chan int
```
简单起见，我们的 `Iter` 结构只处理 `int` 类型。使用 `interface{}` 可以实现一个“泛型”的迭代器，不过在方法中需要通过反射来进行一些类型检查以确保类型安全。比较麻烦，我们就不这么写了，感兴趣的朋友可以参考前文中提到的 Rob Pike 实现的版本。

下面，我们为 `Iter` 添加一些方法，以便于把它用作迭代器。

## 实现 `Map` 和 `Filter`

首先我们来实现著名的 `Map` 和 `Filter`。这个两个方法都在一个 `Iter` 上调用，返回一个新的 `Iter`。下面是 `Map` 的实现：

```go
// iter.go

func (it Iter) Map(fn func(int) int) Iter {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for x := range it {
			ch <- fn(x)
		}
	}()
	return ch
}
```

我们在 `Map` 方法中声明一个新的 channel， 然后发起一个 goroutine，不断地从原先的 `Iter` 中读取元素，经过转化后写到新的 `Iter` 中去。由于 channel 是阻塞的。只有当 `Map` 中新的 `Iter` 读走了一个元素后，原先的 `Iter` 中才会被写入下一个。完美地体现了延迟计算。最后，当原先的 `Iter` 读完（也就 channel 被关闭），关闭新的 channel，以表示迭代到了尾部。

与此类似，`Filter` 的实现如下：

```go
// iter.go

func (it Iter) Filter(pred func(int) bool) Iter {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for x := range it {
			if pred(x) {
				ch <- x
			}
		}
	}()
	return ch
}
```

基本与 `Map` 是一样的，就不赘述了。

## 实现 `Reduce`

`Reduce` 的实现和 `Map` 与 `Filter` 有所不同。因为 `Reduce` 是对所有进行加总，所以每个元素都要被求值。因此 `Reduce` 是无法延迟计算的。如果把 `Map` 和 `Filter` 看作是在 `Iter` 间进行转化的管道的话，`Reduce` 就是这个管道的终点。

具体实现如下：

```go
// iter.go

func (it Iter) Reduce(init int, fn func(int, int) int) int {
	acc := init
	for x := range it {
		acc = fn(acc, x)
	}
	return acc
}
```

其中， `init` 是用于加总的初始值，`fn` 是加总函数，它接受两个参数，第一个是当前为止的加总结果，第二个是当前的元素值，返回添加当前元素后的加总结果。

## 使用 `Range` 创建迭代器

到这里，我们已经实现了经典的集合操作“三剑客”方法：`Map`、`Filter`、`Reduce`。这些方法要么是把一个 `Iter` 转化成另一个 `Iter`，要么是把 `Iter` 加总为一个值。那么最初的 `Iter` 是怎么来的呢？为了让 `Iter` 的创建更加便捷，我们实现一个 `Range(from, to)` 函数。它返回一个包含从 `from` 到 `to` 的整数的 `Iter`。这个函数是管道的起点。

```go
// iter.go

func Range(from, to int) Iter {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := from; i < to; i++ {
			ch <- i
		}
	}()
	return ch
}
```

`Range` 创建一个 `Iter`，把 `from` 到 `to` 的整数依次传入，最后关闭 channel。

## 迭代器的威力：无穷序列 `Seq`

前面提到，迭代器与普通 `for` 循环相比的一大优势就是可以表示无穷集合。为了突显这一点，我们再实现一个 `Seq` 方法。它构建一个递增地返回所有自然数（包括0）的 `Iter`。

```go
// iter.go

func Seq() Iter {
	ch := make(chan int)
	n := 0
	go func() {
		for {
			ch <- n
			n++
		}
	}()
	return ch
}
```

`Seq` 中的 channel 永远不会关闭，因此逻辑上它表示一个无穷的序列。但是因为 channel 的阻塞特性，我们每次只会取出一个元素进行计算。在 `Seq()` 上可以调用 `Map` 和 `Reduce`，因为计算是延迟的。但是如果我们直接在 `Seq()` 上调用 `Reduce` 就会陷入死循环。因为 `Reduce` 不是延迟计算，它会试图取出 `Iter` 里的所有元素，直到 `Iter` 被关闭（而 `Seq()` 永远不会被关闭）。

## 化无穷为有穷：`Take`

为了避免 `Reduce` 陷入死循环，我们需要一种方式把无穷的 `Iter` 转化为有穷的 `Iter`。所以，我们再实现一个 `Take` 方法，它截取一个 `Iter` 的前 n 个元素。

```go
// iter.go

func (it Iter) Take(n int) Iter {
	count := 0
	ch := make(chan int)
	go func() {
		defer close(ch)
		for x := range it {
			if count < n {
				ch <- x
				count++
			} else {
				break
			}
		}
	}()
	return ch
}
```

这样，`Seq().Take(10)`，就可以得到前10个自然数。

为了完整起见，我们还实现一个 `Drop` 方法。和 `Take` 相反，它跳过一个 `Iter` 的前 n 个元素，保留后面的所有元素：

```go
// iter.go

func (it Iter) Drop(n int) Iter {
	count := 0
	ch := make(chan int)
	go func() {
		defer close(ch)
		for x := range it {
			if count < n {
				count++
			} else {
				ch <- x
			}
		}
	}()
	return ch
}
```

最后，为了便于打印，我们实现一个 `Collect` 方法，它把一个 `Iter` 转化为一个 `[]int`：

```go
// iter.go

func (it Iter) Collect() []int {
	var s []int
	for x := range it {
		s = append(s, x)
	}
	return s
}
```

与 `Reduce` 类似，在无穷 `Iter` 上调用 `Collect` 也会导致死循环。


到此为止，我们就完成了所需的所有关于 `Iter` 的方法。接下来我们通过几个具体的例子，展示如何使用这个 `Iter` 类型进行集合操作。

## 例子1：输出正整数的平方

`square` 函数通过组合 `Range` 和 `Map` 来输出前 n 个正整数的平方：

```go
// main.go
package main

func squares(n int) []int {
	return Range(1, n+1).
		Map(func(x int) int { return x * x }).
		Collect()
}

func main() {
    fmt.Printf("20以内正整数的平方：%v\n", squares(20))

    // ...
}

```

运行程序将输出：
```
20以内正整数的平方：[1 4 9 16 25 36 49 64 81 100 121 144 169 196 225 256 289 324 361 400]
```

## 例子2：计算阶乘

利用 `Reduce`，可以计算阶乘：

```go
// main.go

func fac(n int) int {
	return Range(1, n+1).
		Reduce(1, func(acc, cur int) int { return acc * cur })
}

func main() {
    // ...

    fmt.Printf("10的阶乘: %d\n", fac(15))

    // ...
}

```

运行程序将输出：
```
10的阶乘：3628800
```

## 例子3：输出前 n 个质数

最后，我们看一个稍微更加复杂的例子。这个例子利用了无穷 `Iter` 和延迟计算的特点，输出前 n 个质数：

```go
// iter.go

func primes(n int) []int {
	return Seq().
		Drop(2).  // 跳过0和1
		Filter(isPrime).
		Take(n).
		Collect()
}

// helper
func isPrime(n int) bool {
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
    // ...

    fmt.Printf("前100个质数为：%v\n", primes(100))
}

```

运行程序，将输出：
```
前100个质数为：[2 3 5 7 11 13 17 19 23 29 31 37 41 43 47 53 59 61 67 71 73 79 83 89 97 101 103 107 109 113 127 131 137 139 149 151 157 163 167 173 179 181 191 193 197 199 211 223 227 229 233 239 241 251 257 263 269 271 277 281 283 293 307 311 313 317 331 337 347 349 353 359 367 373 379 383 389 397 401 409 419 421 431 433 439 443 449 457 461 463 467 479 487 491 499 503 509 521 523 541]
```

## 总结

本文中我们通过实现一个 `Iter` 迭代器类型，探讨了 Go 语言中 channel 和迭代器的相似性。希望对大家有所启发。这些代码仅仅是作为探讨展示所用，所以忽略了很多必要的边界检查、错误处理等等。也有一些很有用的方法因为篇幅原因我们未能实现， 比如 `First` （返回第一个符合条件的值），`Until` （返回迭代器中第一个符合条件的元素之前的部分），`Any` （检测迭代器中是否有任一元素满足条件），`All` （检测迭代器中是否所有元素都符合某一条件），`Zip` （把两个迭代器交错式地合为一个），如此等等。有兴趣的读者可以自行实现。