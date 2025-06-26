package main

import (
	"fmt"
	"math/rand"
	"time"
)

// It uses generator pattern to generate even and odd numbers

func main() {
	even := printEven()
	odd := printOdd()
	for i := 0; i < 25; i++ {
		fmt.Printf(<-even)
		fmt.Printf(<-odd)
	}
	fmt.Println("You're both boring; I'm leaving.")
}

func printEven() chan string {
	c := make(chan string)
	go func() {
		for i := 0; ; i += 2 {
			c <- fmt.Sprintf("Even: %d\n", i)
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
		}
	}()
	return c
}

func printOdd() chan string {
	c := make(chan string)
	go func() {
		for i := 1; ; i += 2 {
			c <- fmt.Sprintf("Odd: %d\n", i)
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
		}
	}()
	return c
}
