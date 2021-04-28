package main

import (
	"reflect"
	"testing"
)

func TestRange(t *testing.T) {
	from, to := 10, 100
	var expected, actual []int
	for i := from; i < to; i++ {
		expected = append(expected, i)
	}
	it := Range(from, to)
	for x := range it {
		actual = append(actual, x)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Range(%d, %d): expecting %v, got %v", from, to, expected, actual)
	}
}

func TestSeq(t *testing.T) {
	end := 100
	var expected, actual []int
	for i := 0; i < end; i++ {
		expected = append(expected, i)
	}
	it := Seq()
	for x := range it {
		if x == end {
			break
		}
		actual = append(actual, x)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Seq (taking the first %d): expectiing %v, got %v", end, expected, actual)
	}
}

func TestTakeIterLargerThanLimit(t *testing.T) {
	size, limit := 100, 50
	it := makeIter(size)
	var expected, actual []int
	for i := 0; i < limit; i++ {
		expected = append(expected, i)
	}
	for x := range it.Take(limit) {
		actual = append(actual, x)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Take(%d), size = %d: expecting %v, got %v", limit, size, expected, actual)
	}
}

func TestTakeIterSmallerThanLimit(t *testing.T) {
	size, limit := 50, 100
	it := makeIter(size)
	var expected, actual []int
	for i := 0; i < size; i++ {
		expected = append(expected, i)
	}
	for x := range it.Take(limit) {
		actual = append(actual, x)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Take(%d), size = %d: expecting %v, got %v", limit, size, expected, actual)
	}
}

func TestDropIterLargerThanLimit(t *testing.T) {
	size, limit := 100, 50
	it := makeIter(size)
	var expected, actual []int
	for i := limit; i < size; i++ {
		expected = append(expected, i)
	}
	for x := range it.Drop(limit) {
		actual = append(actual, x)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Drop(%d), size = %d: expecting %v, got %v", limit, size, expected, actual)
	}
}

func TestDropIterSmallerThanLimit(t *testing.T) {
	size, limit := 50, 100
	it := makeIter(size)
	var expected, actual []int // expected will remain nil
	for x := range it.Drop(limit) {
		actual = append(actual, x)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Drop(%d), size = %d: expecting %v, got %v", limit, size, expected, actual)
	}
}

func TestCollect(t *testing.T) {
	size := 100
	it := makeIter(size)
	var expected, actual []int
	for i := 0; i < size; i++ {
		expected = append(expected, i)
	}
	actual = it.Collect()

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Drop(): expecting %v, got %v", expected, actual)
	}
}

func TestMap(t *testing.T) {
	size := 10
	it := makeIter(size)
	double := func(x int) int { return x * x }
	var expected, actual []int
	for i := 0; i < size; i++ {
		expected = append(expected, double(i))
	}
	for x := range it.Map(double) {
		actual = append(actual, x)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Map(double): expecting %v, got %v", expected, actual)
	}
}

func TestFilter(t *testing.T) {
	size := 10
	it := makeIter(size)
	isEven := func(x int) bool { return x%2 == 0 }
	var expected, actual []int
	for i := 0; i < size; i++ {
		if isEven(i) {
			expected = append(expected, i)
		}
	}
	for x := range it.Filter(isEven) {
		actual = append(actual, x)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Filter(isEven): expecting %v, got %v", expected, actual)
	}
}

func TestReduce(t *testing.T) {
	size := 10
	it := makeIter(size)
	add := func(acc, cur int) int { return acc + cur }
	var expected, actual int
	for i := 0; i < size; i++ {
		expected += i
	}
	actual = it.Reduce(0, add)

	if expected != actual {
		t.Errorf("Reduce(add): expecting %d, got %d", expected, actual)
	}
}

func makeIter(max int) Iter {
	it := make(Iter)
	go func() {
		defer close(it)
		for i := 0; i < max; i++ {
			it <- i
		}
	}()
	return it
}
