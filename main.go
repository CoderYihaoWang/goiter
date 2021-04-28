package main

import "fmt"

// Run:
// go run iter.go main.go
//
// Test:
// go test iter.go iter_test.go
func main() {
	// print the squares of 1 through 20
	// 打印 1 到 20 的平方
	fmt.Printf("Squares of 1 ~ 20: %v\n", squares(20))

	// print the factorial of 10
	// 打印 10 的阶乘
	fmt.Printf("Factorial of 10: %d\n", fac(10))

	// print the first 100 prime numbers
	// 打印前 100 个质数
	fmt.Printf("The first 100 prime numbers: %v\n", primes(100))
}

// squares of 1 ~ n, inclusive
// 返回 1 ~ n 间整数的平方，包含端点
func squares(n int) []int {
	return Range(1, n+1).
		Map(func(x int) int { return x * x }).
		Collect()
}

// the factorial of positive integer n
// 计算正整数 n 的阶乘。
func fac(n int) int {
	return Range(1, n+1).
		Reduce(1, func(acc, cur int) int { return acc * cur })
}

// first n-th prime numbers
// 返回前 n 个质数
func primes(n int) []int {
	return Seq().
		Drop(2).
		Filter(isPrime).
		Take(n).
		Collect()
}

func isPrime(n int) bool {
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}
