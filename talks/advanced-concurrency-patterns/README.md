https://youtu.be/QDDwwePbDtw?si=FVO41kZIBicjIEx-

# Go supports concurrency
Go supports concurrency in the language, not as a library. That means you have great syntax for dealing with 
starting concurrent events, communicating with them and then reacting to that communication.

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

Datatype Ball keeps track of how many times it has been hit. We are going to pass pointer to the same ball 
back and forth to the Goroutines.

One of the big benefits of having this concurrency managed by the Go runtime is, we can do things like detect 
when the system is deadlocked. So if we never send the ball on the channel, we will do one second sleep and 
then we'll get the fatal error.

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

It says all Goroutines are sleep deadlock. It says all Goroutines are sleep deadlock.

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan receive]:
main.main()
	/tmp/sandbox1880805371/prog.go:19 +0xbb

goroutine 6 [chan receive]:
main.player({0x4b3d9b, 0x4}, 0xc000076070)
	/tmp/sandbox1880805371/prog.go:24 +0x37
created by main.main in goroutine 1
	/tmp/sandbox1880805371/prog.go:14 +0x66

goroutine 7 [chan receive]:
main.player({0x4b3d97, 0x4}, 0xc000076070)
	/tmp/sandbox1880805371/prog.go:24 +0x37
created by main.main in goroutine 1
	/tmp/sandbox1880805371/prog.go:15 +0xa5

```

It will print out the stack traces of the Goroutines that are sitting in the system.

We have our main (Goroutine 1) that's blocked on channel receive. Where we are trying to receive from the 
table. And we have two players main.player corresponding to Goroutine 6 and Goroutine 7, each stuck on 
channel receive.

The set of stacktraces can be useful for other things as well.

# Panic dumps the stacks
A panic in Goroutine is going to dump the stacks. There are number of ways to get the stacktraces. Runtime
package provides it. There's even ways to export it through HTTP handlers. So you can look at, say a
running server and see where it's Goroutine stacks are.

```go 
package main

import (
	"fmt"
	"time"
)

type Ball struct {
	hits int
}

func main() {
	table := make(chan *Ball)
	go player("ping", table)
	go player("pong", table)

	table <- new(Ball)  // game on; toss the Ball
	time.Sleep(1 * time.Second)
	<-table // game over; grab the Ball

    panic("show me the stacks")
}

func player(name string, table chan *Ball) {
	for {
		ball := <-table
		ball.hits++
		fmt.Println(name, ball.hits)
		time.Sleep(100 * time.Millisecond)
		table <- ball
	}
}
```

We used to see similar stack trace earlier so there was a Goroutine leak. But with newer version of Go
we don't see that anymore.

# It's easy to Go, but how to Stop
Long lived program need to cleanup.

Let's look at how to write programs that handle communication, periodic events, and cancellation.

The core's Go's **select** statement: like a **switch**, but the decisison is made based on the ability to
communicate.

```go 
select {
case xc <- x:
    // sent x on xc
case y := <- yc:
    // received y from yc
}
```

Select blocks until one of it's cases proceed. Each case is a communication.

## Example: Feed Reader
Where do we start?

I know they have something to do with RSS.

### Find an RSS client
We can open up this website godoc.org, it provides automatically generated documentation for various Go 
projects on BitBucket, Github, Google hosting and launchpad. And so if you search for RSS here, you will get
a number of Go packages. We get their tag lines here that says what this package is about.

A number of these are just providing Go data structures that correspond to the RSS data types. But some of
them actually provide clients. That's what we really want.

We can click through one of these packages e.g. https://github.com/jteeuwen/go-pkg-rss
and get more information on what the package owner has described and this is all powered by godoc.org.

So we choose a package and started looking at it's interface

```go
// Fetch fetches Items for uri and returns the time when the next 
// fetch should be attempted. On failure, Fetch retruns and error.
func Fetch(uri string) (items []Item, next time.Time, err error)

type Item struct {
    Title, Channel, GUID string // a subset of RSS fields.
}
```

What it provides is a fetch function, it takes a uri from which to fetch the feed, and it returns a slice of
Items and a time which is the time next time we should try a fetch on this particular feed. Or if there's a
failure, it'll return an error value.

But this is not what we want, we want stream of items that we can then build UI on top of. So what we really
want is channel of items.

```go
<- chan Item
```

It's receive only channel. We just want to suck on that channel and do interesting things with the items that
we get. Also multiple subscription not just one. So let's take a look in more detail about what we want.
Here's what we have:

```go
type Fetcher interface {
    Fetch() (items []Item, next time.Time, err error)
}

func Fetch(domain string) Fetcher {...} // fetches Items from domain
```

We can call fetch periodically to get items and we're going to put in an interface so it makes it easy to
provide a fake implementation for this talk and you can imagine using it for testing.

We have our function, which given our domain, a blogger domain, is going to give us a fetcher implementation.
And that's where we are starting.

```go 
type Subscription interface {
    Updates() <- chan Item  // Stream of items
    Close() error           // Shuts down the stream
}

func Subscribe(fetcher Fetcher) Subscription {...}  // Converts Fetches to a stream
func Merge(subs ...Subscription) Subscription {...} // Merges several streams
```

What we want to build is something more like a Subscription that gives us a Channel on which we can receive 
these items and gives us a way to say we're no longer interested unsubscribe. So wr're going to close that
subscription and it's going to give us an error if there was any problems fetching that stream. And this gets
to that whole cancellation point. We want to be able to cleanup when we're no longer interested in something.

We have ```Subscribe``` function which given a fetcher transforms it into a subscription. It takes this idea
of periodically fetching and gives us this stream. 

And then we have ```Merge``` which takes a variadic list of subscriptions, any number of subscriptions, and
gives us a merged one, one whose series of items are merged from the various domains.

Here is an example that puts it together. This is using a fake fetcher.

```go
func main() {
    // Subscribe to some feeds, and create a merged update stream.
    merged := Merge(
        Subscribe(Fetch("blog.golang.org")),
        Subscribe(Fetch("googleblog.blogspot.com")),
        Subscribe(Fetch("googledevelopers.blogspot.com"))
    )

    // Close the subscriptions after some time.
    time.AfterFunc(3 * time.Second, func() {
        fmt.Println("closed:", merged.Close())
    })

    // Print the stream
    for it := range merged.Updates() {
        fmt.Println(it.Channel, it.Title)
    }

    panic("show me the stacks")
}
```

Panic at the end is the same, it's the same techniques that we have used before. It just shows us that all
that's hanging around is the main Goroutine and a syscall Goroutine managed by the runtime.

There's nothing with regards to subscription that's still hanging around. So we've cleanedup is the point.

```go
func Subscribe(fetcher Fetcher) Subscription { // Converts Fetches to a stream
	s := &subscriptionImpl{
		fetcher: fetcher,
		updates: make(chan Item), // for updates
	}
	go s.loop()
	return s
}

```
The idea behind subscribe is to translate this periodic fetch into a continuous stream of items. The way 
that's going to work is we're going to have an un-exporting data type that implements ```Subscription``` and
that's ```subscriptionImpl```.

```subscriptionImpl``` is going to contain a ```fetcher``` and it's going to have this channel of items it
exposes to the client. And when we create new subscription we are going to initialize our subtype and we're
going to start a Goroutine called ```loop``` and this is the key bit.

This is going to implement that periodic fetching, but also handle events like closing and delivering items
to the user.

## Implementing Subscription
To implement ```Subscription``` interface, define **Updates** and **Close**

```go
func (s *subscriptionImpl) Updates() <- chan Item {
    return s.updates
}

func (s *subscriptionImpl) Close() error {
    // TODO: make loop exit
    // TODO: find out about any error
    return err
}
```

Note that in ```Updates``` we are returning receive only channel. So the client of the subscription is only going tobe able to receive. They won't be able to try to send on that channel.

And ```Close``` has two jobs, it's going to make that ```loop``` Goroutine exit. It has to find out about 
any fetch error that happened while we're trying to get that subscription.

## What does loop do?
- Periodically call ```Fetch```
- send fetched items on the **Updates** channel.
- exit when ```Close``` is called, reporting any error.

This doesn't sound too hard. So here is the implementation that you might just try first.

```go
func (s *subscriptionImpl) loop() {
	for {
		if s.closed {
			close(s.updates)
			return
		}
		items, next, err := s.fetcher.Fetch()
		if err != nil {
			s.err = err
			time.Sleep(10 * time.Second)
			continue
		}
		for _, item := range items {
			s.updates <- item
		}
		if now := time.Now(); next.After(now) {
			time.Sleep(next.Sub(now))
		}
	}
}
```

If we run this it seems to kind of work. We seem to be getting items. And it looks like it even might
have exited cleanly in this case. But turns out this code is very buggy.

## Bug 1: unsynchronized access to s.closed/s.err
```s.closed``` & ```s.err``` is being accessed by two Goroutines with no synchronization, this is bad.
This means that this is a data race.

You can see particularly in case of error, which is an interface value, you can see partially written types,
if there really are because these are executing in shared address space.

You notice that this race is sort of by inspection, which is not very satisfying way to find data races.

### Race Detector
Luckily Go 1.1 onwards we have a new race detector. So we can run our program as below to find out.

```sh
go run -race main.go
```

So the race detector is really great way in a program that is doing a lot of concurrency, that's dealing
with lots of Goroutines and a lot of data, to discover when you are not synchronizing properly.

## Bug 2: time.Sleep may keep loop running
This is really a resource leak of an insidious sort, which is our loop is just sleeping when it has nothing
else to do and what that means is it's not responsive when it's time to close a subscription.

Now that might not be a big deal, it's just one Goroutine hanging around, we know it's going to exit 
eventually it'll see closed. But what if this next time is tomorrow? Some feeds are updated fairly rarely.
So this could be hanging around for quite a long time. We want to be in this case that as soon as the user
calls ```Close``` we clean up, because there's no reason for this to hang around.

## Bug 3: loop may block forever on s.updates
It's the hardest bug to catch. This send ```s.updates <- item``` will block until there's a Goroutine ready 
to receive the value. So what happens when the client of the subscription calls ```Close``` and then stops
using the subscription. There is no reason they should keep receiving off of it. Well this send will block
indefinitely. There's no one to receive it, nothing's going to happen. This ```loop``` goroutine is going to
stay hanging around in the system forever.

Luckily we have a way to fix all of these problems by refactoring the way our loop works.

## Solution
Change the body of ```loop``` to a ```select``` with three cases:
- ```Close``` was called.
- It's time to call ```Fetch```
- send an item on ```s.updates```

To consider multiple things that could happen at the same time, when one happens, take action on it, then
consider them all again.

### Structure: for-select loop
```loop``` runs in its own goroutine. ```select``` lets ```loop``` avoid blocking indefinitely in any one
state.

```go
func (s *subscriptionImpl) loop() {
    ... declare mutable state ...
    for {
        ... setup channels for cases ...
        select {
        case <- c1: 
            ... read/write state ...
        case c2 <- x:
            ... read/write state ...
        case y := <- c3:
            ... read/write state ...
        }
    }
}
```

The cases interact via local state in ```loop```.

And the key here is that there is no data races because this is a single Goroutine. It's a straight line code.
It's much more algorithmic than you might be used to in dealing with threads and event-driven style.

We will see how powerful this is and the way we resolve these errors and then build on this structure.

So we are going Case by case.

### Case 1: Close
We need to transform that ```closed``` boolean and that ```err``` value in communication. And we are going
to do this by introducing a channel in our subscriber implementation called ```closing```.

```go
type subscriberImpl struct {
    closing chan chan error
}
```

It's a channel on which you pass a channel. What's going on here?
This is a request/response structure. So the loop you can think of as a little server that's listening for
requests on this closing channel, in particular request to ```Close```. And when we call ```Close``` we're
going to say, hey loop, please close and we're going to wait until the loop acknowledges that by sending us 
back the error if there's been an error on fetch, or nil if there's been no error.

And this request/response structure can be built up quite a bit.

```Close``` asks loop to exit and waits for a response. 
```go
func (s *subscriptionImpl) Close() error {
    errc := make(chan error)
    s.closing <- errc
    return <-errc
}
```
So the code here is that ```Close``` now creates the channel, it wants to receive an errors, sends that
to the loop then waits for the error to come back.


```loop``` handles ```Close``` by replying with the ```Fetch``` error and exiting.

```go
var err error   // set when Fecth fails
for {
    select {
    case errc := <-s.closing:
        errc <- err
        close(s.updates)    // tells receiver we're done
        return
    }
}
```
In loop, we have this is just a snippet of it, our first select case, which is if we receive a value on 
closing, deliver the error to the listener, close updates channel, and then return. And this ```err``` 
variable is the one that fetch is going to write to and fetch is next.

### Case 2: Fetch
This is the big case 

```go
var pending []Item  // appeneded by fetch; consumed by send
var next time.Time  // Initially January 1, year 0
var err error
for {
    var fetchDelay time.Duration    // Initially 0 (no delay)
    if now := time.Now(); next.After(now) {
        fetchDelay = next.Sub(now)
    }
    startFetch := time.After(fetchDelay)

    select {
    case <- startFetch:
        var fetched []Item
        fetched, next, err = s.fetcher.Fetch()
        if err != nil {
            next = time.Now().Add(10 * time.Second)
            break
        }
        pending = append(pending, fetched...)
    }
}
```

State here is, what's returned by fetch and shared with the other cases. In particular, we have a set of items
that we fetched previously and that we want to deliver to the client. We have this ```next``` time that we
need to run our next fetch and we have the error from a failed fetch.

At the top of our loop, we have to calculate when is our next fetch. This ```fetchDelay```, initially it's
going to be 0. So our first fetch is going to be immediately.

```startFetch``` is a channel, we use the ```time.After``` function in the time package to create a channel
that will deliver a value exactly after ```fetchDelay``` elapses. And this second select case is going to be
able to proceed after that time has elapsed.

In this case we are going to go ahead and fetch our items, we'll run ```s.fetcher.Fetch```. If there is an
error we'll schedule our next fetch to be 10 seconds from now. You can imagine backing off there.

And if we succeeded we'll append the fetched items to our pending queue.

### Case 3: Send
Our last case is going to then deliver the items in those pending queue to the client of the type. This seems
to be pretty easy, we should be able to just take the first item off pending and deliver it on the updates
channel and then advance the slice.

```go
var pending []Item  // appended by fetch; consumed by send
for {
    select {
    case s.updates <- pending[0]:
        pending = pending[1:]
    }
}
```

Unfortunately it crashes exactly when pending is empty and this is because the way select is evaluated. It
evaluates the expressions in the communication and it evaluated ```pending[0]``` when it runs select and if
it's empty, that's going to be a panic. We are going to fix it using another technique.

#### Select and nil channels
It's going to combine two properties of the Go language that are somewhat interesting. First is that if you 
send or receive on a nil channel, it blocks. It doesn't panic it just blocks. And we know that select will
never select a blocking case. It will only select cases that proceed.

```go
func main() {
    a, b := make(chan string), make(chan string)
    go func() { a <- "a" }()
    go func() { b <- "b" }()

    // flip a coin and set one of it to nil
    if rand.Intn(2) == 0 {
        a = nil
        fmt.Println("nil a")
    } else {
        b = nil
        fmt.Println("nil b")
    }

    select {
        case s := <- a:
            fmt.Println("got", s)
        case s := <- b:
            fmt.Println("got", s)
    }
}
```

If we put this together, we can use the fact that we can set a channel to nil to disble a case that we don't
need this time around. And this replaces a lot more complicated logic where you have multiple different
selects depending on what can happen.

So below is our modified code to fix the third case.

```go
var pending []Item  // appened by fetch; consumed by send
for {
    var first Item
    var updates chan Item
    if len(pending) > 0 {
        first = pending[0]
        updates = s.updates // enable send case
    }

    select {
    case updates <- first:
        pending = pending[1:]
    }
}
```

The way we are going to fix it now is, in our select case we're only going to attempt to send if we've got 
value to send. So we're going to create an updates channel that's nil and we're going to set it to non-nil
exactly when you have something to send, then we're going to set our first item to ```pending[0]```.

Now we can put all these cases together.

```go
select {
case errc := <-s.closing:
    errc <- err
    close(s.updates)    // tells receiver we're done
    return

case <- startFetch:
    var fetched []Item
    fetched, next, err = s.fetcher.Fetch()
    if err != nil {
        next = time.Now().Add(10 * time.Second)
        break
    }
    pending = append(pending, fetched...)

case updates <- first:
    pending = pending[1:]
}
```

The key thing is:
The cases interact via ```err```, ```next``` and ```pending```. No locks, no condition variables, no callback.

But we are doing fairly complex updates to our state because we are dealing with lots of things going on.
And yet, this is still responsive. When it's time to send something to the user or the user's no longer
interested, we cleanup very rapidly.

Now that we have this structure we can make a number of improvements. It's much easier to see what's going
on in this data type.

## Issue 1: Fetch may return duplicates
We can fix this trivially now inside loop by just keeping track of all the ones we've already seen because
we don't want to deliver duplicate items on our stream.

```go
var pending []Item
var next time.Time
var err error
var seen = make(map[string]bool)    // set of item.GUIDs
```

Instead of just appending all the items we've

```go
case <-startFetch:
    var fetched []Item
    fecthed, next, err = s.fetcher.Fetch()
    if err != nil {
        next = time.Now().Add(10 * time.Second)
        break
    }
    for _, item := range fetched {
        if !seen[item.GUID] {
            pending = append(pending, item)
            seen[item.GUID] = true
        }
    }
```

## Issue 2: Pending queue grows without bound
Set of pending items can grow without bound because our downstream receiver may be distracted, may not be
keeping up with the items we're offering. You can think of this as ```back pressure```.

If the receiver is doing something else, we don't want to keep acquiring more and more pending items and
allocating arbitrary amount of memory. So what we want to do here is bound, when the reciever is slow, we
want to stop fetching new items. We can use that nil channel trick again.

Remember that startFetch schedule told us when the next time we should fetch. If we leave that to nil, we
won't run another fetch. 

So we are going to decide, in this case that the maximum number of pending items we have is 10.

```go
const maxPending = 10
```

We're only going to schedule the next fetch if we have room in our pending queue.

```go
var fetchDelay = time.Duration
if now := time.Now(); next.After(now) {
    fetchDelay = next.Sub(now)
}
var startFetch <- chan time.Time
if len(pending) < maxPending {
    startFetch = time.After(fetchDelay) // enable fetch case
}
```

We could have instead drop older items from the head of ```pending```. It's upto us how we want to deal 
with it.

## Issue 3: Loop blocks on Fetch
Fetches are doing IO, they are talking to remote servers, and they could take a long time, and so this 
particular call ```s.fetcher.Fetch``` may block and we may not want to block our loop on these fetches.
We want to remain responsive. If the server is taking 10 seconds to respond and the user calls ```Close```,
we want to be able to ditch and move on.

And the way we are going to do is we're going to move fetch to it's own Goroutine. But now we have to figure 
out when that fetch finishes, we need to then resynchronize back with what loop is doing.

We'll do this by introducing another case into our select and another channel for running fetches to
communicate with us. Here is how that works.

### Fix: Run Fetch Asynchronously
First we need a channel on which select can send its results. We're going to define a ```fetchResult```
datatype, which is a struct containing the fetched items, the next time and error.

```go
type fetchResult struct {
    fetched []Item
    next    time.Time
    err     error
}
```

We have this ```fetchDone``` channel, which is channel of fetchResults. And our invariant, this is non-nil
exactly when fetch is running.

```go
var fetchDone chan fetchResult  // if non-nil, Fetch is running
```

So we want to start a fetch when fetchDone is nil. That's our initial state.

```go
var startFetch <- chan time.Time
if fetchDone == nil && len(pending) < maxPending {
    startFetch = time.After(fetchDelay) // enable fetch case
}
```

So we'll go ahead and start a fetch and the ```startFetch``` case is going to set our ```fetchDone``` channel
and start a Goroutine that runs our fetch.

```go
select {
case <- startFetch:
    fetchDone = make(chan fetchResult, 1)
    go func() {
        fetched, next, err := s.fetcher.Fetch()
        fetchDone <- fetchResult{fetched, next, err}
    }()
case result := <- fetchDone:
    fetchDone = nil
    // use result.fetched, result.next, result.err
}
```

We made our loop not block on the IO and we still synchronize properly with all the other events that can
happen in this subscription.

# Implemented Subscribe
So we've implemented subscribe. This code is responsive to various events in the environment, it cleansup
everything and handles cancellation properly and the structure that we introduced made it relatively easy
to read and change the code.
We used these three key-techniques:
- ```for-select``` loop. That allows you to consider all of the things that could be happening to your type.
- service channel, reply channels (chan chan error). You can build quite a lot on that. Where I think of 
Goroutines in GO, you can have them acting as servers and clients on a little distributed system that you own.
And you can use these channels to exchange data. And you build a lot of the same structures, you see 
distributed system within your program.
- Finally we had this technique of setting channels to nil and select cases to turn cases on and off. That 
really helps you simplify some of the logic when you are dealing with type like this.

# Conculsion
Concurrent programming can be tricky.

Go makes it easier:
- Channels convey data, timer events, cancellation signals.
- Goroutines serialize access to local mutable state 
- Stack traces and deadlock detector
- race detector.

# Links
- Go concurrency patterns: talks.golang.org/2012/concurrency.slide
- Concurrency is not parallelism: golang.org/s/concurrency-is-not-parallelism
- Share memory by communicating: golang.org/doc/codewalk/sharemem
- Go tour (learn Go in your browser): tour.golang.org

Q: Is there any tooling available to detect leaking Goroutines?
A: For now we are using stack traces to detect leaking Goroutines and frankly that can be very effective,
particularly in tests and isolation. It would be interesting to see if we can automate it. It's little bit
hard to really detect when code is done executing. ```Blocking Profiles``` is in Go1.1 and you can get a 
graph of what's blocked.

While writing any concurrent code remember below tagline:
**Share memory by communicating, not communicate by sharing memory.** We were sharing memory without 
synchronization, we fixed that by converting it to communication.

Q: One of the popular approach on JVM is to use **Actors** and the way those approach concurrency is by kind 
of holding their hands up and saying, no one can touch this message queue except the object that it's assigned
to. It seems Goroutines and Channels are similar to the approach, except that the channels can be touched by
external sources.
A: Difference here is that you can use Goroutines and channels to build something Actor like, which is sortof
what we did with the four select group. You can also use them to do a variety of other things. It's strength
of the language. Serializing access to mutable state, which is what Actors do as well. So do critical sections.
