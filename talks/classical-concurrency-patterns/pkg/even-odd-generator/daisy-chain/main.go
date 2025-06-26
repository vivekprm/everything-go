package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

func f(left, right chan int) {
	left <- <-right + 1
}

func main() {
	n := 10
	leftmost := make(chan int)
	left := leftmost
	right := leftmost
	for i := 0; i < n; i++ {
		right = make(chan int)
		go f(left, right)
		left = right
	}
	// Get the number of running Goroutines
	numGoroutines := runtime.NumGoroutine()
	fmt.Printf("Goroutines: %d\n", numGoroutines)
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
	go func(c chan int) {
		c <- 1
	}(right)
	fmt.Println(<-leftmost)
}
