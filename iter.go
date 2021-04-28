package main

// Iter demostrates how to use a Go channels to mimic iterators.
// Note that this program is for demostration purpose only,
// to simplify things, we only use int as the type of elements,
// and many necessary boundary checkings and error handlings in the methods are omitted.
//
// Iter 类型展示了怎样使用 Go 语言的 channel 来模拟迭代器。
// 提示：本程序仅用作探索展示使用。简便起见，我们只支持 int 作为元素类型，
// 并且在下面的方法中，许多必要的边界检查和错误处理都被略过了。
type Iter <-chan int

// Map creates a new Iter whose elements are projected from those of the original Iter
// by applying the fn argument.
//
// Map 方法生成一个新的迭代器，并使用参数 fn 将旧迭代器中的元素映射到新迭代器中。
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

// Filter creates a new Iter which only contains the elements from the original Iter that
// satisfies the pred argument.
//
// Filter 方法生成一个新的迭代器，只保留旧迭代器中满足 pred 条件的元素。
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

// Reduce aggregates the elements of the Iter by applying the fn argument.
// The initial value is specified by the init argument.
// DO NOT call Reduce on an infinite Iter, otherwise the program will enter an infinite loop.
//
// Reduce 方法对迭代器中的元素使用 fn 参数进行加总。init 参数是用于加总的初始值。
// 不要在无穷迭代器上调用此方法，否则会导致死循环。
func (it Iter) Reduce(init int, fn func(int, int) int) int {
	acc := init
	for x := range it {
		acc = fn(acc, x)
	}
	return acc
}

// Range generates an Iter containing integers [from, to)
//
// Range 方法生成一个包含 [from, to) 区间中整数的迭代器。
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

// Seq creates an infinite Iter containing integers starting from 0
//
// Seq 方法生成包含从0开始的整数的无穷迭代器。
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

// Take creates an Iter that only contains the first at most n elements of the original Iter.
//
// Take 方法创建一个新的迭代器，只包含原先迭代器中的最多前 n 个元素。
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

// Drop creates an Iter that skips over the first at most n elements of the original Iter.
//
// Drop 方法创建一个新的迭代器，跳过原先迭代器中的最多前 n 个元素。
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

// Collect turns an Iter to a slice.
// DO NOT call this method on an infinite Iter, or it results in an infinite loop.
//
// Collect 方法将一个迭代器转化成一个 slice。
// 不要在无穷迭代器上调用此方法，否则会导致死循环。
func (it Iter) Collect() []int {
	var s []int
	for x := range it {
		s = append(s, x)
	}
	return s
}
