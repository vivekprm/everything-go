package main

import (
	"fmt"
	"math/rand"
	"time"
)

type number struct {
	num  int
	wait chan bool
}

func main() {
	limit := 50
	even := generateEvenSequence(limit)
	odd := generateOddSequence(limit)
	seq := generate(even, odd)

	for {
		select {
		case s := <-seq:
			fmt.Println(s.num)
			s.wait <- true
		case <-time.After(100 * time.Millisecond):
			fmt.Println("timed out")
			return
		}
	}
}

func generateEvenSequence(limit int) chan number {
	waitForIt := make(chan bool)
	c := make(chan number)
	go func() {
		for i := 0; i <= limit; i += 2 {
			c <- number{num: i, wait: waitForIt}
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
			<-waitForIt
		}
	}()
	return c
}

func generateOddSequence(limit int) chan number {
	waitForIt := make(chan bool)
	c := make(chan number)
	go func() {
		for i := 1; i <= limit; i += 2 {
			c <- number{num: i, wait: waitForIt}
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
			<-waitForIt
		}
	}()
	return c
}

func generate(c1, c2 chan number) chan number {
	c := make(chan number)
	go func() {
		for {
			select {
			case even := <-c1:
				c <- even
			case odd := <-c2:
				c <- odd
			}
		}
	}()
	return c
}
