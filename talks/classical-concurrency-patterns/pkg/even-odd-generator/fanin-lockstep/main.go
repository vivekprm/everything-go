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

	for msg := range seq {
		fmt.Println(msg.num)
		msg.wait <- true
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

func generate(chs ...chan number) chan number {
	c := make(chan number)
	for i := range chs {
		ch := chs[i]
		go func() {
			for {
				c <- <-ch
			}
		}()
	}
	return c
}
