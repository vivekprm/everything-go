package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Message struct {
	str  string
	wait chan bool
}

func fanin(inputs ...<-chan Message) <-chan Message {
	out := make(chan Message)
	for i := range inputs {
		input := inputs[i]
		go func() {
			for {
				out <- <-input
			}
		}()
	}
	return out
}

func main() {
	c := fanin(boring("Joe"), boring("Ann"))
	for i := 0; i < 5; i++ {
		msg1 := <-c
		fmt.Println(msg1.str)
		msg2 := <-c
		fmt.Println(msg2.str)
		msg1.wait <- true
		msg2.wait <- true
	}
	fmt.Println("You're both boring; I am leaving.")
}

func boring(msg string) <-chan Message {
	waitForIt := make(chan bool) // shared between all messages
	c := make(chan Message)
	go func() {
		for i := 0; ; i++ {
			c <- Message{fmt.Sprintf("%s %d", msg, i), waitForIt}
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
			<-waitForIt
		}
	}()
	return c
}
