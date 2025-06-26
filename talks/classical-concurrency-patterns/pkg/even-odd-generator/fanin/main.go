package main

import "fmt"

func main() {
	limit := 50
	even := generateEvenSequence(limit)
	odd := generateOddSequence(limit)
	seq := generate(even, odd)

	for i := range seq {
		fmt.Println(i)
	}
}

func generateEvenSequence(limit int) chan string {
	c := make(chan string)
	go func() {
		for i := 0; i <= limit; i += 2 {
			c <- fmt.Sprintf("Even: %d\n", i)
		}
	}()
	return c
}

func generateOddSequence(limit int) chan string {
	c := make(chan string)
	go func() {
		for i := 1; i <= limit; i += 2 {
			c <- fmt.Sprintf("Odd: %d\n", i)
		}
	}()
	return c
}

func generate(c1, c2 chan string) chan string {
	c := make(chan string)
	go func() {
		for {
			c <- <-c1
		}
	}()
	go func() {
		for {
			c <- <-c2
		}
	}()
	return c
}
