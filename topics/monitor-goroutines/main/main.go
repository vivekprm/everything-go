package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	var wg *sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	noOfWorker := 5
	ch := make(chan int, 100)
	done := make(chan int, 100)

	wg.Add(1)
	go monitorGoroutines(ctx, wg)

	for x := range noOfWorker {
		wg.Add(1)
		go worker(wg, x, ch, done)
	}

	for x := range 100 {
		ch <- x
	}
	close(ch)
}

func monitorGoroutines(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("context is done")
			return
		default:
			time.Sleep(time.Second)
			fmt.Printf("Number of goroutines: %d\n", runtime.NumGoroutine())
		}
	}
}
