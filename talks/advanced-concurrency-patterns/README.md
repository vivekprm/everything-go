https://youtu.be/QDDwwePbDtw?si=FVO41kZIBicjIEx-

# Go supports concurrency
Go supports concurrency in the language, not as a library. That means you have great syntax for dealing with starting concurrent events,
communicating with them and then reacting to that communication.

# Goroutines & Channels
Goroutines are independently executing functions in the same address space.

Channels are typed values that allow Goroutines to synchronize and exchange information.

```go 
c := make(chan int)
go func() { c <- 3 }()
n := <- c
```

## Example: ping-pong
```go 
type Ball struct {
    hits int
}

func main() {
    table := make(chan *Ball)
    go player("ping", table)
    go player("pong", table)

    table <- new(Ball)  // game on; toss the Ball
    time.Sleep(1 * time.Second)
    <- table // game over; grab the Ball
}

func player(name string, table chan *Ball) {
    for {
        ball := <- table 
        ball.hits++
        fmt.Println(name, ball.hits)
        time.Sleep(100 * time.Millisecond)
        table <- ball
    }
}
```

Datatype Ball keeps track of how many times it has been hit. We are going to pass pointer to the same ball back and forth to the 
Goroutines.

One of the big benefits of having this concurrency managed by the Go runtime is, we can do things like detect when the system is
deadlocked. So if we never send the ball on the channel, we will do one second sleep and then we'll get the fatal error.

```go 
type Ball struct {
    hits int
}

func main() {
    table := make(chan *Ball)
    go player("ping", table)
    go player("pong", table)

    // table <- new(Ball)  // game on; toss the Ball
    time.Sleep(1 * time.Second)
    <- table // game over; grab the Ball
}

func player(name string, table chan *Ball) {
    for {
        ball := <- table 
        ball.hits++
        fmt.Println(name, ball.hits)
        time.Sleep(100 * time.Millisecond)
        table <- ball
    }
}
```
