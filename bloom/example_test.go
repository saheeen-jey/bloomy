package bloom

import "fmt"

func ExampleFilter_Add() {
	filter := New(100, 3)
	filter.Add([]byte("apple"))

	fmt.Println(filter.Test([]byte("apple")))
	// Output: true
}

func ExampleFilter_Union() {
	left := New(100, 3)
	right := New(100, 3)
	left.Add([]byte("apple"))
	right.Add([]byte("banana"))

	_ = left.Union(right)
	fmt.Println(left.Test([]byte("banana")))
	// Output: true
}
