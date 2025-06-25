# What is Concurrency?
Concurrency is the composition of independently executing computations.

Concurrency is a way to structure software, particularly as a way to write clean code that interacts well with the real world.

## Concurrency Is Not parallelism
- Concurrency is not parallelism, although it enables parallelism.
- If you have only one processor, your program can still be concurrent but it can't be parallel.
- On the other hand, a well written concurrent program might run efficiently in parallel on a multiprocessor.
- See tinyurl.com/goconcnotpar for more on that distinction.

# History
- Rooted in, Hoare's CSP (Communicating Sequential Processes) in 1978 paper.

# Distinction
- Go is the latest on the Newsqueak-Alef-Limbo languages branch, distinguished by first class channels.
- Erlang is closer to the original CSP, where you can communicate to a process by name rather than over a channel.
- The models are equivalent but express things differently.
- Rough analogy: Writing to a file by name (process, Erlang) vs. writing to a file descriptor (channel, Go)

# Basic Examples
## A Boring Function
We need an example to show the interesting properties of the concurrency primitives.
To avoid distraction, we make it a boring example.

```go
func boring (msg string) {
    for i := 0; ; i++ {
        fmt.Println(msg, i)
        time.Sleep(time.Second)
    }
}
```

Let's make it realistic by making it sleep for random intervals:

```go
func boring (msg string) {
    for i := 0; ; i++ {
        fmt.Println(msg, i)
        time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
    }
}

```

## Ignoring It 
- The ```go``` statement runs the function as usual, but doesn't make the caller wait.
- It launches a goroutine.
- The functionality is analogous to the & on the end of a shell command.

```go
package main 

import (
    "fmt"
    "math/rand"
    "time"
)

func main() {
    go boring("boring!")
}
```

So as the main returns program exits, so the goroutine is killed immediately.

So let's do couple of things.

```go 
func main() {
    go boring("boring!")
    fmt.Println("I'm listening.")
    time.Sleep(2 * time.Second)
    fmt.Println("You're boring; I'm leaving.")
}
```

# What are Goroutines?
It's an independently executing function, launched by a ```go``` statement. So in Go concurrency is a composition of independently
executing Goroutines.

It has it's own call stack, which grows and shrinks as required. So unlike some threading libraries, where you have to say how big the
stack is, it's never an issue in Go. It will be made as big as it needs to be, and if the stack grows the system will take care of the
stack growth for us. 

They start out very small. So it's very cheap and practical to have thousands or even tens of thousands Goroutines running.

It's not a thread. However, for the point of view of understanding them, it's not misleading to think of a Goroutine as just an 
extremely cheap thread. What happens in the runtime is Goroutines are multiplexed onto threads that are created as needed in order 
to make sure no Goroutine ever blocks.

# Proper concurrent program
- Our boring example cheated: the main function couldn't see the output from the other Goroutine.
- It was just printed on the screen, where we pretended we saw a conversation.
- Real conversations require communication.

To do real communication there is a concept of **channels** in Go. A **Channel** in Go provides a connection between two Goroutines,
allowing them to communicate.

```go 
// Declaring and initializing
var c chan int
c = make(chan int)
// or
c := make(chan int)
```

```go 
// sending on a channel
c <- 1
```

```go 
// Receiving from a channel.
// The arrow indicates the direction of data flow.
value = <- c
```

Now let's use the channel to do something.

A channel connects the main and boring goroutines so they can communicate.

```go 
func main() {
    c := make(chan string)
    go boring("boring!", c)
    for i := 0; i < 5; i++ {
        fmt.Printf("You say: %q\n", <- c) // Receive expression is just a value
    }
    fmt.Println("You're boring; I'm leaving.")
}
```

```go 
func boring(msg string, c chan string) {
    for i := 0; ; i++ {
        c <- fmt.Sprintf("%s %d", msg, i) // Expression to be sent can be any suitable value
        time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
    }
}
```

So when the main function executes ```<- c```, it will wait for a value to be sent.

Similarly when the boring function executes ```c <- value```, it waits for a receiver to be ready.

A sender and receiver must both be ready to play their part in the communication. Otherwise we wait until they are.

Thus *channels both communicate and synchronize*.

- **Note for experts**: Go channels can also be created with a buffer.
- Buffering removes synchronization.
- Buffering makes them more like Erlang's mailboxes.
- Buffered channels can be important for some problems but they are more subtle to reason about.
- We won't need them today.

# The Go Approach
*Don't communicate by sharing memory, share memory by communicating*. In other words, you don't have some blob of memory and then 
put locks and mutexes and condition variables around it to protect it from parallel access. Instead, you actually use the channel to 
pass the data back and forth between the Goroutines.

# Concurrency Patterns
## Generator: function that returns a channel
Channels are first-class values, just like string or integers.

```go 
c := boring("boring!")  // function returning a channel

for i := 0; i < 5; i++ {
    fmt.Printf("You say: %q\n", <- c)
}
fmt.Println("You're boring; I'm leaving.")
```

```go 
func boring(msg string) <-chan string { // returns receive only channel of strings
    c := make(chan string)
    go func() { // We launch the Goroutine from inside the function
        for i := 0; ; i++ {
            c <- fmt.Sprintf("%s %d", msg, i)
            time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
        }
    }()
    return c // Return the channel to the caller
}
```

### Channels as a handle on a service
Our boring function returns a channel that lets us communicate with the boring service it provides. We can have more instances of the
service.

```go 
func main() {
    joe := boring("Joe")
    ann := boring("Ann")
    for i := 0; i < 5; i++ {
        fmt.Println(<- joe)
        fmt.Println(<- ann)
    }
    fmt.Println("You're both boring; I'm leaving.")
}
```

Inside the for loop we are reading a value from Joe and Ann nd because of the synchronization nature of the channels, the two guys are 
taking turns, not only in printing the values out, but also in executing them. Because if Ann is ready to send a value and Joe hasn't 
done that yet, Ann will still be blocked, waiting to deliver the value to the main.

That's annoying because may be Ann is more talkative than Joe and doesn't want to wait around.

So we can get around that by writing a fan-in function or a **Multiplexer**

## Multiplexer
These programs make Joe and Ann count in lockstep.
We can instead use a fan-in function to let whosoever is ready to talk.

```go
// It's also a generator pattern. Takes two channels as input and returns another channel as output.
func fanin(input1, input2 <-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for {
			c <- <-input1
		}
	}()

	go func() {
		for {
			c <- <-input2
		}
	}()

	return c
}
```

```go 
func main() {
	c := fanin(boring("Joe"), boring("Ann"))
	for i := 0; i < 10; i++ {
		fmt.Println(<-c)
	}
	fmt.Println("You are both boring; I am leaving.")
}
```

Here in fanin we actually stitch the two guys together and construct a single channel, from which we can receive from both of them.

![image](https://github.com/user-attachments/assets/4c0c7434-27aa-4700-bffd-776a3e50dc13)
 

What if for some reason we don't want it and we wanted to have them totally lockstep and synchronous.

## Restoring Sequencing
To have them totally lockstep, remember these channels in Go are first class values in Go. That means we can pass a channel on a 
channel. So we can send inside a channel another channel to be used for the answer to comeback.

To do this what we do is we construct a message structure that includes the message that we want to print and we include inside it 
another channel that's what we call wait channel and that's like **signaler** and the guy will block on wait channel until the 
person says ok I want you to go ahead.

```go 
type Message struct {
    str string
    wait chan bool
}
```

Each speaker must wait for a go-ahead.

```go 
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
```

Now in boring function now we have this waitForIt channel and then everybody blocks waiting for signal to advance.

# Select statement
It is a control structure, somewhat like a switch, that lets you control the behaviour of your program based on what communications
we are able to proceed any moment.

The **select** statement is really sort of a key part of why concurrency is built into Go as feature of the language, rather than
just a library. It's harder to control structures that are part of library. It's much easier this way.

The select statement provides another way to handle multiple channels.
It's like a switch but each case is a communication:
- All channels are evaluated.
- Selection blocks until one communication can proceed, which then does.
- If multiple can proceed, select chooses pseudo-randomly.
- A default clause, if present, executes immediately if no channel is ready.

```go 
select {
case v1 := <- c1:
    fmt.Printf("received %v from c1\n", v1)
case v2 := <- c2:
    fmt.Printf("received %v from c2\n", v2)
case c3 <- 23:
    fmt.Printf("Sent %v to c3\n", 23)
default:
    fmt.Printf("no one was ready to communicate.\n")
}
```

## Fanin Again
Rewrite our original fanin function. Only one goroutine is needed.

Old:
```go 
func fanin(input1, input2 <- chan string) <- chan string {
    c := make(chan string)
    go func() { for{ c <- <- input1 } }()
    go func() { for{ c <- <- input2 } }()
    return c
}
```

New:

```go 
func fanin(input1, input2 <- chan string) <- chan string {
    c := make(chan string)
    go func() {
        for {
            select {
            case s := <- input1:
                c <- s
            case s := <- input2:
                c <- s
            }
        }
    }()
    return c
}
``` 

## Timeout Using Select 
The ```time.After``` function returns a channel that blocks for a specified duration. After the interval, the channel delivers the
current time, once:

```go 
func main() {
    c := boring("Joe")
    for {
        select {
        case s := <- c:
            fmt.Println(s)
        case <- time.After(1 * time.Second):
            fmt.Println("You're too slow.")
            return 
        }
    }
}
```

Now we can do that another way, We might decide instead of having a conversation where each message is at most one second, we might
just want a total time elapsed.

## Timout for whole conversation using Select
Create the timer once, outside the loop, to timeout the enitre conversation. (In above program we had a timeout for each message).

```go 
func main() {
    c := boring("Joe")
    timeout := time.After(5 * time.Second)
    for {
        select {
        case s := <- c:
            fmt.Println(s)
        case <- timeout:
            fmt.Println("You talk too much.")
            return 
        }
    }
}
```

## Quit Channel
We can turn this around and tell Joe to stop when we're tired of listening to him.

```go 
quit := make(chan bool)
c := boring("Joe", quit)
for i := rand.Intn(10); i >= 0; i-- {
    fmt.Println(<-c)
}
quit <- true
```

Now listen for value at quit channel and leave once received.

```go 
select {
case c <- fmt.Sprintf("%s: %d", msg, i):
    // do nothing
case <- quit:
    return
}
```

### Receive on Quit channel
How do we know it's finished? Wait for it to tell us it's done: receive on the quit channel.

```go 
quit := make(chan string)
c := boring("Joe", quit)
for i := rand.Intn(10); i>= 0; i-- {
    fmt.Println(<- c)
}
quit <- "Bye!"
fmt.Println("Joe says: %q\n", <- quit)
```

```go 
select {
case c <- fmt.Sprintf("%s: %d\n", msg, i):
    // do nothing
case <- quit:
    cleanup()
    quit <- "See you!"
    return
}
```

So here we do round trip communication, first we send "Bye!" on the quit channel, we receive on the quit channel in select statement.
We do the cleanup then send back again of quit channel. Which prints "Joe says: ".

## Daisy-chain
Speaking of round trips, we can also make this crazy by having a ridiculously long sequence of these things, one talking to another
one. So think of it like this:

pic 

You've got a whole bunch of gophers who want to do a Chinese Whispers game, although I think Chinese Whispers with megaphones might 
sort of make it little weired.

To make it interesting gophers receive from the right and then we make it a channel of integers, so I'm going to add 1. And the reason
for that is, that lets us count the number of steps, so the distortion in the Chinese Whispers game is we add 1 to the value.

```go 
func f(left, right chan int) {
    left <- 1 + <- right
}

func main() {
    const n = 100000
    leftmost := make(chan int)
    right := leftmost
    left := leftmost
    for i := 0; i < n; i++ {
        right = make(chan int)
        go f(left, right)
        left = right
    }
    go func(c chan int) {
        c <- 1
    } (right)
    fmt.Println(<-leftmost)
}
```

So what this code does basically construct the diagram above. Everybody is waiting for first thing to be sent and so we launch the
value in the first channel and then wait for it to comeout of the leftmost guy. So for fun we are going to run 100000 gophers.

Go was designed for building system software. And now we want to talk now about how we use these ideas to construct the kind of
software that we care about.

# System Software 
Now we are going to design Google Search engine.

## Example: Google Search
Q: What does Google search do?
A: Given a query, return a page of search results (and some ads).

Q: How do we get the search results?
A: Send the query to Web search, Image search, YouTube, Maps, News, etc. Then mix the results.

How do we implement this?

## Google Search: A fake framework
We can siumlate the search function, much as we simulated conversation before.

```go 
var (
    Web = fakeSearch("web")
    Image = fakeSearch("image")
    Video = fakeSearch("video")
)

type Search func(query string) Result

func fakeSearch(kind string) Search {
    return func(query string) Result {
        time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
        return Result(fmt.Sprintf("%s result for %q\n", kind, query))
    }
}
```

## Google Search: Test the framework
```go 
func main() {
    rand.Seed(time.Now().UnixNano())
    start := time.Now()
    results := Google("golang")
    elapsed := time.Since(start)
    fmt.Println(results)
    fmt.Println(elapsed)
}
```

## Google Search 1.0
The Google function takes a query and returns a slice of Results (which are just strings).
Google invokes Web, Image & Video searches serially, appending them to Results slice.

```go 
func Google(query string) (results []Result) {
	results = append(results, Web(query))
	results = append(results, Image(query))
	results = append(results, Video(query))
	return results
}
```

## Google Search 2.0
In version 1.0 we launch web search wait for the results and then launch Image search wait for the results and then launch video
search. Instead we can launch them concurrently.

```go 
func Google(query string) (results []Result) {
    c := make(chan Result)
    go func() {
        c <- Web(query)
    }()
    go func() {
        c <- Image(query)
    }()
    go func() {
        c <- Video(query)
    }()

    for i := 0; i < 3; i++ {
        result := <- c
        results = append(results, result)
    }
    return
}
```

Now we are just waiting for the slowest search result.

Now this is a parallel program with multiple backends running but we don't have any mutexes or locks or condition variables. 

Sometimes server can take really really long time and can be very very slow.

## Google Search 2.1
Don't wait for slow servers. No locks. No condition variables.  No callbacks 

Sometimes server can take really really long time and can be very very slow.

## Google Search 2.1
Don't wait for slow servers. No locks. No condition variables.  No callbacks.

```go 
func Google(query string) (results []Result) {
	c := make(chan Result)
	go func() {
		c <- Web(query)
	}()
	go func() {
		c <- Image(query)
	}()
	go func() {
		c <- Video(query)
	}()
    
	timeout := time.After(80 * time.Millisecond)
	for i := 0; i < 3; i++ {
		select {
		case result := <-c:
			results = append(results, result)
		case <-timeout:
			fmt.Println("timeout")
			return
		}
	}
	return
}
```

However, timing out a communication is kind of annoying. What if the server really is going to take a long time? It's kind of a shame
to throw it on the floor. So now we add replication.

## Avoid Timeout
Q: How do we avoid discarding results from slow servers?
A: Replicate the servers. Send requests to multiple replicas, and use the first response.

```go 
func First(query string, replicas ...Search) Result {
    c := make(chan Result)
    searchReplica := func(i int) {
        c <- replicas[i](query)
    }
    for i := range replicas {
        go searchReplica(i)
    }
    return <- c
}
```

### Using the First Function
```go 
func main() {
    rand.Seed(time.Now().UnixNano())
    start := time.Now()
    result := First("golang", fakeSearch("replica 1"), fakeSearch("replica 2"))
    elapsed := time.Since(start)
    fmt.Println(result)
    fmt.Println(elapsed)
}
```

Let's stitch all this magic together in Google-Search-3.0

## Google Search 3.0
```go
func Google(query string) (results []Result) {
	c := make(chan Result)
	go func() {
		c <- First(query, Web1, Web2)
	}()
	go func() {
		c <- First(query, Image1, Image2)
	}()
	go func() {
		c <- First(query, Video1, Video2)
	}()

	timeout := time.After(65 * time.Millisecond)

	for i := 0; i < 3; i++ {
		select {
		case result := <-c:
			results = append(results, result)
		case <-timeout:
			fmt.Println("timed out")
			return
		}
	}
	return
}
```

Now you can see most of the time we are able to get all the search results without timeout. And we have used no locks, no conditional
variables, no callbacks (so it's very different from e.g. using nodejs).

This program is fairly easy to understand. More importantly, the individual elements of the program are all just straight forward 
sequential code and we are composing their independent execution to give us the behaviour of the total server.

# Summary
In just a few simple transformations we used Go's concurrency primitives to convert a
- slow
- sequential
- failure-sensitive

program into one that is
- fast
- concurrent
- replicated
- robust

# More party tricks
There are endless ways to use these tools, many presented elsewhere.

Chatroulette toy:
    tinyurl.com/gochatroulette

Load balancer:
    tinyurl.com/goloadbalancer

Concurrent prime sieve:
    tinyurl.com/gosieve

Concurrent power series:
    tinyurl.com/gopowerseries

# Don't overdo it
They're fun to play with, but don't overuse these ideas.

Goroutines and Channels are big ideas. They are tools for program construction.

But sometimes all you need is a reference counter.

Go has "sync" and "sync/atomic" packages that provide mutexes, condition variables, etc. They provide tools for similar problems.

Often, these things will work together to solve a bigger problem.

Always use the right tool for the Job.

# Conclusion
Goroutines and Channels make it easy to express complex operations dealing with:
- Multiple inputs
- Multiple outputs
- timeouts
- failures

And they're fun to use.

# Links
- Go tour (learn Go in your browser)
    https://go.dev/tour/welcome/1
- Package documentation
    https://pkg.go.dev/std
- Articles galore
    https://go.dev/doc/
- Concurrency is not parallelism
    tinyurl.com/goconcnotpar

Question: If you think about Goroutines as external services, you're thnking about basically integration testing. What are the best practices?
Answer: It's actually a non-issue because of the way language works. Look at the below generator example:

```go
c := boring("boring!")
for i := 0; i < 5; i++ {
	fmt.Printf("You say: %q\n", <-c)
}
fmt.Println("You are boring; I am leaving.")
```

Total interface to boring is a channel, nowhere does this function here know what that channel has behind it. So we can easily mock
this service. So whole idea of a channel is to hide what's behind it. Because it's a first class citizen in the language.

Question: Tools to do static analysis of concurrent program?
Answer: Thread Sanitizer from Google. From communication perspective tools like Spin that Gerald Holzmann wrote.

Question: Is there a way, if you create a channel, to determine how many readers and writers there are for a given channel?
Answer: There is a built-in function you can use to see how many values are in a buffered channel. For non buffered channel there is
no way at the moment to find out how many readers and writers there are. One of the reasns for that is any such question is inherently
unsafe. Because if you care how many readers and wrtiers there are, then that'e becasue you're going to do some computation based on 
that. But that computation might be wrong.





