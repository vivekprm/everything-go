https://youtu.be/sKvFXAkQqXY?si=odJskQr_6jdycKH_
https://github.com/MarioCarrion/videos/tree/7a00f3d7fc1573985d4403d122a230a2554acb93/2021/10/01

# Background Job
A process in charge of doing some work behind the scenes.

It might be initialized by another "parent" process.

A Gorouting launching N Goroutines.

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Process ID", os.Getpid())

	listenForWork()

	<-waitToExit()

	fmt.Println("exiting")
}

func listenForWork() {
	const workersN int = 5

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM)

	//-

	workersC := make(chan struct{}, workersN)

	// 1) Listen for messages to process
	go func() {
		for {
			<-sc

			workersC <- struct{}{} // 2) Send to processing channel
		}
	}()

	go func() {
		var workers int

		for range workersC { // 3) Wait for messages to process
			workerID := (workers % workersN) + 1
			workers++

			fmt.Printf("%d<-\n", workerID)

			go func() { // 4) Process messages
				doWork(workerID)
			}()
		}
	}()
}

func waitToExit() <-chan struct{} {
	runC := make(chan struct{}, 1)

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, os.Interrupt)

	go func() {
		defer close(runC)

		<-sc
	}()

	return runC
}

func doWork(id int) {
	fmt.Printf("<-%d starting\n", id)

	time.Sleep(3 * time.Second)

	fmt.Printf("<-%d completed\n", id)
}
```

To trigger the jobs we can run below shell script:

```sh
for i in {1..5}; do kill -TERM 15498; done
```

We can't trigger more than 5 jobs as we reached the buffer capacity of the buffered channel.

Here Goroutines will not exit cleanly as there are no way to tell goroutines to exit.

# Advanced Example
Here we are going to define a buffer as well as number of channels/number of workers.

For graceful shutdown we are listening for interrupt signal (CTRL + C) 