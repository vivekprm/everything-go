# Rethinking Classical Concurrency Patterns
https://youtu.be/5zXAHh5tJqQ?si=oML56RU8rtT5ch7Y

Keep below two principles in mind:
- Start Goroutines when you have concurrent work.
- Share by communicating.

# Asynchronous APIs
Concurrency is not parallelism. Concurrency is not asynchronicity either. An **Asynchronous API** is the one that returns
to the calling function early. An asynchronous program is not necessarily concurrent. A program could call an asynchronous 
function and then sit idle waiting for the results. Some poorly written asynchronous programs do exactly that. They make a 
sequential chain of calls that each return early wait for result and then start the next call.

## Asynchronous Callback: API
Programmers coming to Go from certain languages such as JavaScript sometimes start with asynchronous callbacks. 

```go 
// Fetch immediately returns, then fetches the item and
// invokes f in a Goroutine when the item is available.
// If the item doesn't exist,
// Fetch invokes f on the zero item.
func Fetch(name string, f func(Item)) {
    go func() {
        [...]
        f(item)
    }()
}
```

The problem with asynchronous callbacks are well described already. So don't use this pattern.

## FUTURE: API
```Future``` pattern instead of returning the result, the function returns a proxy object that allows the caller to wait for the 
result at some later point. You may also know ```Futures``` by the name ```async``` and ```await``` that have built-in support
for the pattern.

```go 
// Fetch immediately returns a channel, then fetches the requested item
// and sends it on the channel.
// If the item doesn't exist,
// Fetch closes the channel without sending.
func Fetch(name string) <- chan Item {
    c := make(chan Item, 1)
    go func() {
        [...]
        c <- item
    }()
    return c
}
```

Usual Go analog for a future is a single element buffered channel that receives a single value and often starts a Goroutine 
to compute that value. It's not exactly the conventional future pattern since we can only receive the value from the channel once
but the channel pattern seems to be a lot more common than the function based alternative.

### FUTURE: CALL SITE
Callers of a Future based APIs setup the work then retrieve the result. If they retrieve the results too early the program executes
sequentially instead of concurrently.

Yes
```go 
a := Fetch("a")
b := Fetch("b")
consume(<-a, <-b)
```

No 
```go 
a := <- Fetch("a")
b := <- Fetch("b")
consume(a, b)
```

## PRODUCER - CONSUMER QUEUE: API
A producer concumer queue also returns a channel but the channel receives any number of results and is typically unbuffered.

```go 
// Glob finds all items with names matching pattern
// and sends them on the returned channel.
// It closes the channel when all items have been sent.
func Glob(pattern string) <- chan Item {
    c := make(chan Item)
    go func() {
        defer close(c)
        for [...] {
            [...]
            c <- item
        }
    }()
    return c
}
```

### Producer - Consumer Queue: Call SITE
The call site is a range loop rather than a single receive operation 

```go 
for item := range Glob("[ab]*") {
    [...]
}
```

# Classical Benefit
Now that we know what an asynchronous API looks like let's examine the reasons we might want to use them. 

- **Responsiveness**: Avoid blocking UI and Network threads
Most other languages don't multiplex across OS threads and kernel schedulers can be unpredictable. So some popular
languages and frameworks keep of the UI or Netwrok logic on a single thread if that thread makes a call that blocks 
for too long the UI becomes Choppy or a network latency spikes. Since calls to asynchronous API is by definition don't 
block they help keep the single threaded programs responsive.

- **Efficiency: Reduce idle threads
On some platforms OS threads are or atleast historically have been expensive. Languages that don't multiplex over threads
can use asynchronous API to keep threads busy reducing the total number of threads and context switches needed to run 
the program.

These first two benifits don't apply in Go. The **runtime** manages threads for us so there is no single UI or network thread to 
block and we don't have to touch the kernel to switch Goroutines.

The runtime also resizes and reallocates thread stack if needed, so Goroutine stacks can be very small and they don't need to 
fragment the address space with guard pages. Today a Goroutine stack starts around 2KB which is half the size of the smallest
amd64 page.

- **Efficiency: Reclaim stack frames 
An asynchronous call may allow the caller to return from arbitrily many frames of the stack that frres up the memory
containing those stack frames for other uses and allows the Go runtime to collect any other allocations that are only reachable 
from those frames.

*Each variable in Go exists as long as there are references to it. **The storage location chosen by the implementation is 
irrelevant** to the semantics of the language.*

Sometime reclaiming stack frames is an optimization but sometimes it is not. Any reference that escapes its frame must be 
allocated in the heap and heap allocations are more expensive in terms of CPU, memory and cache. Furthermore the compiler
can already prune out any stack allocations that it knows are unreachable. It can move large allocations to the heap and 
the grabage collector can ignore dead references. Finally the benefit of this optimization depends on the specific call site.

If the caller doesn't have a lot of data on the stack in the first place then making the call asynchronous won't help much.
When we take all that into account, asynchronicity as an optimization is subtle it requires careful benchmarks for the impact
on specific callers and the impact may change or even reverse from one version of the runtime to the next. It's not the sort 
of optimization we want to build a stable API around.

# GO BENIFIT
So a final benifit of asynchronous API is really does apply in GO.
When an asynchronous function returns the caller can immediately make further calls to start other concurrent work. 
Concurrency can be especially important for network RPCs where the CPU cost of a call is very low compared to it's latency.

Unfortunately that benefit comes at the cost of making the caller side of the API much less clear. Let's look at some examples:

Suppose we come across an asynchronous call while we are debugging or doing a code reivew what can we infer about it from the 
call site?

```go 
a := Fetch("a")
b := Fetch("b")
if err := [...] {
    return err
}
consume(<-a, <-b)
```

What happens if we return early without waiting for the futures to complete? How long will they continue using resources? Might
we start fetches faster than we can retire them and run out of memory? Will fetch keep using the passed in context after it's 
returned. If so what happens if we cancel it and then try to read from the channel? Will we receive a zero value, some other 
sentinel value, will be block?

If we return without draining the channel from Glob will we leak a Goroutine that's sending to it?
```go 
for result := range Glob("[ab]*") {
    if err := [...] {
        return err
    }
}
```

Will Glob keep using the passed in context as we iterate over the results? If so what happens if we cancel it? Will we still 
get results when if ever will the channel be closed in that case?

```go 
for result := range Glob(ctx, "[ab]*") {
    [...]
}
```

These asynchronous APIs raise a lot of questions and to answer those questions we would have to go digging around in the 
documentation. If the answers are even there. So let's rethink this pattern.

How can we get the benefit of asynchronicity without this ambiguity?
We are using Goroutines to implement these asynchronous APIs but what is a Goroutine anyway? A Goroutine is the execution of 
a function, if we don't have another function to execute a Goroutine adds complexity without benefit.

Benefit of asynchronicity is that it allows the caller to initiate other work but how do we know that the caller even has any 
other work? Functions like fetch and Glob shouldn't need to know what other work their callers may be doing that's not their
job.

# Asynchronous == Synchronous
In languages without threads or Goroutines Asynchronous API is our Viral. If we can't execute function calls concurrently 
any function that may be concurrent must be asynchronous.

In contrast in Go it's very easy to wrap an asynchronous API to make it Synchronous or vice-versa.

```go 
func Async(x In) (<-chan Out) {
    c := make(chan Out, 1)
    go func() {
        c <- Synchronous(x)
    }()
    return c
}
```

```go 
func Synchronous(x In) Out {
    c := Async(x)
    return <- c
}
```

We can write the clearer API and adapt it as needed at the call site.

# Add Concurrency on the Caller Side of the API
If we keep the API synchronous, we may need to add the Concurrency at the call site.

## CALLER-SIDE CONCURRENCY: SYNCHRONOUS API
Consider the synchronous version of our Fetch function. The cancellation and error behaviour is so obvious from the function
signature that we don't need extra documentation for it.

```go 
// Fetch returns the requested item
func Fetch(context.Context, name string) (Item, error) {
    [...]
}
```

Now the caller can use whatever pattern they like to add Concurrency. In manyc cases they won't even need to go through channels.
So the questions about channel usage won't even arise.

```go 
var a, b Item 
g, ctx := errgroup.WithContext(ctx)
g.Go(func() (err error){
    a, err = Fetch(ctx, "a")
    return err
})
g.Go(func() (err error){
    b, err := Fetch(ctx, "b")
    return err
})
err := g.Wait()
[...]
consume(a, b)
```
Here we are using ```golang.org/x/sync/errgroup``` package and writing the results directly into local variables.

# Make Concurrency an Internal Detail
As long as we present a simple synchronous API to the caller they don't need to care how many concurrent calls it's implementation
makes. For example consider a synchronous version of our Glob function.

```go 
// Glob finds all items with names matching pattern
func Glob(ctx context.Context, pattern string) ([]Item, error) {
    [...]
}
```

Internally it can fetch all of it's item concurrently and stream them to a channel but the caller doesn't need to know that.
And because the channel is local to the function we can see both the sender and the receiver locally that makes the answers
to our channel question obvious. Since the send is unconditional the receive loop must drain the channel.

```go 
func Glob([...]) ([]Item, err) {
    [...]   // Find matching names
    c := make(chan Item)
    g, ctx := errgroup.WithContext(ctx)
    for _, name := range names {
        name := name
        g.Go(func() error{
            item, err := Fetch(ctx, name)
            if err == nil {
                c <- item
            }
            return err
        })
    }
    go func() {
        err = g.Wait()
        close(c)
    }()
    var items []Item
    for item := range c {
        items = append(items, item)
    }
    if err != nil {
        return nil, err
    }
    return items, nil
}
```

In case of the error the error variable is set and the channel is still closed.

# Concurrency is not Asynchronicity
In Go synchronous and asynchronous APIs are equally expressive. We can call synchronous APIs concurrently and they're 
clearer at the call site. We don't need to pay the cost of asynchronicity to get the benefits of Concurrency.

# Condition variables
Condition variables are next classical pattern and are part of larger Concurrency pattern called **Monitors**. But the
phrase condition variable appears in GO's standard library whereas **Monitor** in this sense doesn't.

Concept of **Monitors** dates to 1973 and Condition variables to 1974. So this is fairly old pattern. First quick refresher:

Let' look at a simple example. An unbounded queue of items a Go condition variable must be associated with a mutex or another
syncedout locker.

```go 
type Queue struct {
    mu sync.Mutex 
    items []Item
    itemAdded sync.Cond
}

func NewQueue() *Queue {
    q := new(Queue)
    q.itemAdded.L = &q.mu 
    return q
}
```

The two basic operations on condition variables are ```wait``` and ```signal```.
```wait``` atomically unlocks the mutex and suspends the calling Goroutine. ```Singal``` wakes up a waiting Goroutine which 
then relocks the mutex before proceeding.

## Condition Variable: Wait & Signal
```go 
func (q *Queue) Get() Item {
    q.mu.Lock()
    defer q.mu.Unlock()
    for len(q.items) == 0 {
        q.itemAdded.Wait()
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item
}

func (q *Queue) Put(item Item) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.items = append(q.items, item)
    q.itemAdded.Singal()
}
```

In our Queue we can use ```wait``` to block on the availability on enqueued items and signal to indicate when another item
has been added.

## Condition Variable: Boradcast
```go 
type Queue struct {
    [...]
    closed bool
}

func (q *Queue) Close() {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.closed = true
    q.cond.Broadcast()
}
```

Boradcast wakes up all waiting Goroutines instead of just one. Broadcast is usually for events that effects all waiters. Such as 
marking the end of the Queue.

However, it is sometime used to wakeup some waiters when we don't know exactly which are eligible.

```go 
func (q [...]) GetMany(n int) []Item {
    q.mu.Lock()
    defer q.mu.Unlock()
    for len(q.items) < n {
        q.itemAdded.Wait()
    }
    items := q.items[:n:n]
    q.items = q.items[n:]
    return items
}

func (q *Queue) Put(item Item) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.items = append(q.items, item)
    q.itemAdded.Broadcast()
}
```

Here we have changes Get to GetMany. After a put one of the waiting GetMany calls may be ready to completed but put has no way
of knowing which oneto wake, so it must wake all of them.

Condition variables have lots of usecases but downside is similar to all of them. So will discuss that one by one.

### Spurious Wakeups
For events that aren't really global broadcast may wakeup too many waiters. For example one call to ```Put``` wakes up of the
```GetMany``` callers eventhough at most one of them will be able to complete. Even ```Signal``` can result is Spurious wakeups.
It ```Put``` used ```Signal``` instead of ```Broadcast```, it could wakeup a caller that is not yet ready instead of one that is.
If it does that repeatedly it could strand items in the queue without corresponding wakeups. If we're very careful we can minimize
or avoid these Spurious wakeups. But that generally adds even more complexity and subtlety to the code. 

### Forgotten Signals
And if we prune out the spurious singals too agressively we risk going too far and dropping some that are actually necessary.
And since the condition variable decouples the signal from the data it's easy to add some new code to update the data and forget
to signal the condition.

### Starvation
Even if we don't forget a signal if the waiters are not uniform the pickier ones can starve. Suppose that we have one caller to 
```GetMany``` 3000 and another caller executing ```GetMany``` 3 in a tight loop. The two waiters will be about equally likely
to wakeup but the GetMany 3 caller will be able to consume 3 items every three calls. Whereas GetMany 3000 won't have enough ready.
The queue will remain drained and the larger call will block forever. If we happen to notice this starvation problem ahead of time 
we could add an explicit wait queue to avoid starvation but that again makes the code more complex.

### Unresponsive cancellation
The whole point of condition variables is to put a Goroutine to sleep while we wait for something to happen but while we're 
waiting for that condition, we may miss some other event that we ought to notice too. For example the caller might decide that
they don't want to wait that long and cancel the passed in context. Expecting us to notice and return more or less immediately.

Unfortunately condition variables only let us wait for events associated with their own mutex. So we can't select on a condition
and a cancellation at the same time even if the caller cancels our call we'll block until the next time the condition is signaled.

Fundamentally condition variables rely on communicating by shared memory they signal that a change has occurred but they leave  
it upto the signaled Goroutine to check other shared variables to figureout what changed. On the otherhand the Go approach is:

**Share by communicating**.

# Share By communicating
Lets look at the usecases of condition variables and rethink them in terms of communication, perhaps we will spot a pattern.

A signal or a broadcast on a condition variable tells the waiters that something has changed often that something is the 
availability of a shared resource such as a connection or a buffer in a pool.

So let's look at a concrete example our resources for this example will be netcons in a pool and we'll start with condition
variable version for reference.

```go 
type Pool struct {
    mu              sync.Mutex
    cond            sync.Cond
    numConns, limit int 
    idle            []net.Conn
}

func NewPool(limit int) *Pool {
    p := &Pool{limit: limit}
    p.cond.L = &p.mu 
    return p 
}

func (p *Pool) Release(c net.Conn) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.idle = append(p.idle, c)
    p.cond.Signal()
}

func (p *Pool) Hijack(c net.Conn) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.numConns--
    p.cond.Signal()
}
```

We got a limit on total number of connections plus a pool of idle connections and a condition variables that tells us when the
set of connections changes. When we are done with the connection, we can either release it back into the idel pool or hijack it 
so that it no longer counts against the limit.

```go 
func (p *Pool) Acquire() (net.Conn, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    for len(p.idle) == 0 && p.numConns > p.limit {
        p.cond.Wait()
    }
    if len(p.idle) > 0 {
        c := p.idle[len(p.idle) - 1]
        p.idle = p.idle[:len(p.idle) - 1]
        return c, nil
    }
    c, err := dial()
    if err != nil {
        p.numConns++
    }
    return c, err
}
```

To acquire a connection, we wait until we have an idle connection to reuse or are under the limit.

Now let's rethink, let's share the resources by communicating the resource. **Resource limits** are resources too.
In particular an available slot toward the limit is a thing that we can consume. Effective Go even have a hint for that.
It mentions

**A buffered channel can be used like a semaphore. The capacity of the channel buffer limits the number of simultaneous calls
to process.**

So we will have a channel for the limit tokens and one for the idle connections resources. Ascend on the semaphore channel will
communicate that we have consumed a slot toward the limit and the idle channel will communicate the actual connections as they 
are idle. 

```go 
type Pool struct {
    sem chan tokens
    idle chan net.Conn
}

type token struct{}

func NewPool(limit int) *Pool {
    sem := make(chan token, limit)
    idle  := make(chan net.Conn, limit)
    return &Pool{sem, idle}
}

func (p *Pool) Release(c net.Conn) {
    p.idle <- c
}

func (p *Pool) Hijack(c net.Conn) {
    <-p.sem
}
```

Now release and Hijack have become trivial. **Release** literally puts the connection back into the Pool. **Hijack** releases
a token from the semaphore. They have dropped from 4 line bodies to one line each. Instead of locking and storing the resource, 
signaling and unlocking, they simply communicate the resource.

If we really wanted to we could use a single channel for this instead of two. We could use nil net.Conn to represent permission
to create a new connection. Personally I think code is clearer with separate channels.

Acquire ends up a lot simpler too. Even cancellation is just one more case in the select.

```go 
func (p *Pool) Acquire(ctx context.Context) (net.Conn, error) {
    select {
    case conn := <-p.idle:
        return conn, nil
    case p.sem <- token{}:
        conn, err := dial()
        if err != nil {
            <-p.sem
        }
        return conn, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

Conditions can also indicate the presence of new data for processing. Let's go back to our queue example.

```go 
func (q *Queue) Get() Item {
    q.mu.Lock()
    defer q.mu.Unlock()
    for len(q.items) == 0 {
        q.itemAdded.Wait()
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item
}

func (q *Queue) Put(item Item) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.items = append(q.items, item)
    q.itemAdded.Singal()
}
```

For the single item Get and Put, a signal indicates the availability of an item of data. While in ```GetMany``` version it 
indicates potential availability of an item that some other goroutine may have already consumed. 

```go 
func (q [...]) GetMany(n int) []Item {
    q.mu.Lock()
    defer q.mu.Unlock()
    for len(q.items) < n {
        q.itemAdded.Wait()
    }
    items := q.items[:n:n]
    q.items = q.items[n:]
    return items
}

func (q *Queue) Put(item Item) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.items = append(q.items, item)
    q.itemAdded.Broadcast()
}
```

That imprecise targetting is the cause of both **Spurious wakeups** and **Starvation**. To avoid spurious wakeups we should 
signal only the Goroutine that will actually consume the data. But if we know which Goroutine will consume the data, we may
as well send the data along too. Sending the data makes it much easier to see whether the signal is spurious. If we resend 
the exact same data to the same receiver or if we call that explicitly ignore the channel receive for example by executing
a continue in a range loop then we probably didn't need to send it in the first place.
Sending the data also makes signal harder to forget will very likely notice if we compute data and then don't send it 
anywhere although we do still have to be careful to send it to all interested receivers.

## Metadata are data too
So how do we identify the right receivers? 
The information about who needs which data is also data, we can communicate that too.

We'll start with single item get. Here we need two channels, one to communicate the items and another to communicate whether
any items even exist. Both will have a buffer size of 1. 

The ```items channel``` functions like a ```mutex``` while the ```empty channel``` is like a ```one token semaphore```.

```go 
type Queue struct {
    items chan []items  // non-empty slices only
    empty chan bool     // holds true if queue is empty.
}

func NewQueue() *Queue {
    items := make(chan []Item, 1)
    empty := make(chan bool, 1)
    empty <- true
    return &Queue{items, empty}
}
```

This time we really need the two separate channels. ```Put``` needs to know when there are no items so that it can start new slice.
But ```Get``` wants only non-empty items.

```go 
func (q *Queue) Get() Item {
    items := <-q.items

    item := items[0]
    items = items[1:]

    if len(items) == 0 {
        q.empty <- true
    } else {
        q.items <- items
    }
    return item
}

func (q *Queue) Put(item Item) {
    var items []Item
    select {
    case items = <-q.items:
    case <-q.empty:
    }
    items = append(items, item)
    q.items <- items
}
```

To support cancellation in Get, all we have to do is move that initial channel receive into a select statement.

```go 
func (q *Queue) Get() Item {
    var items []Item
    select {
    case <- ctx.Done():
        return 0, ctx.Err()
    case items = <-q.items:
    }

    item := items[0]
    items = items[1:]

    if len(items) == 0 {
        q.empty <- true
    } else {
        q.items <- items
    }
    return item
}
```

We don't need to select on the sends at the end because we know that they won't block. When we received the items, we also 
received the information that our Goroutine owns those items. So tha't Get with one item, what about GetMany?

To figure out whether we should wake GetMany caller, we need to know how many items it wants then we need a channel on which
we can send those items to that particular caller. We'll put the items and the metadata together in one Queue state struct
and just for good measure we'll share that state by communicating it to a channel with a one element buffer functions much
like a selectable mutex.

```go 
type waiter struct {
    n int
    c chan []Item
}

type state struct {
    items []Item
    wait  []waiter
}

type Queue struct {
    s chan state
}

func NewQueue() *Queue {
    s := make(chan state, 1)
    s <- state{}
    return &Queue{s}
}
```

To get a run of items we first check the current state for sufficient items if there're not enough we add an entry to the 
metadata. To put an item to the queue we append it to the current state and then check the metadata to see whether that 
makes enough items to send to the next waiter. When we don't have enough items left we'll stop sending items and send back
the updated state.

```go 
func ([...]) GetMany(n int) []Item {
    s := <-q.s 
    if len(s.wait) == 0 && len(s.items) >= n {
        items := s.items[:n:n]
        s.items = s.items[n:]
        q.s <- s 
        return items
    }
    c := make(chan []Item)
    s.wait = append(s.wait, waiter{n, c})
    q.s <- s 
    return <-c
}

func (q *Queue) Put(item Item) {
    s := <-q.s 
    s.items = append(s.items, item)
    for len(s.wait) > 0{
        w := s.wait[0]
        if len(s.items) < w.n {
            break
        }
        w.c <- s.items[:w.n:w.n]
        s.items = s.items[w.n:]
        s.wait = s.wait[1:]
    }
    q.s <- s
}
```

Since all of this communication occurs on channels it's possible to plumbing cancellation here too. Do it as an exercise.

Broadcast on a condition may signal a transition from one state to another for example it may indicate that the program 
has finished loading its initial configuration or that a communication stream has been terminated. One simple state transition
is from Busy to Idle.

Using condition variables, we need to store the state explicitly. You might think that we would only need to store the current 
state the busy boolean but that turns out to be a very subtle decision. If a wait idle looped only until it saw a non-busy state
it would be possible to transition from busy to idle and back before a wait idle got a chance to check and we would miss short
idle events.

```go 
type Idler struct {
    mu      sync.Mutex 
    idle    sync.Cond 
    busy    bool
    idles   int64
}

func NewIdler() *Idler {
    i := new(Idler)
    i.idle.L = &i.mu 
    return i
}

func (i *Idler) AwaitIdle() {
    i.mu.Lock()
    defer i.mu.Unlock()
    idles = i.idles
    for i.busy && idles == i.idles {
        i.idle.Wait()
    }
}

func (i *Idler) SetBusy(b bool) {
    i.mu.Lock()
    defer i.mu.Unlock()
    wasBusy := i.busy
    i.busy = b 
    if wasBusy && !i.busy {
        i.idles++
        i.idle.Broadcast()
    }
}
```

GO's condition variables unlike pthread condition variables don't have spurious wakeups, so in theory we could return from a wait
idle unconditionally after the first wait call. However it's also common for condition based code to intentionally oversignal. For
example to work around an undiagnosed deadlock. So to avoid using subtle problems later it's best to keep the code robust to spurious
wakeups. Instead we can track the cumulative count of events and wait until either we catch the idle event in progress or observe
it's effect on the counter.

# Share Completion By Completing communication
We can avoid the double state transition race entirely by communicating the transition instead of signaling it and we can plumbin 
cancellation tto boot. We can broadcast a state transition by closing a channel. There's a nice symmetry to that, a state transition
marks the Completion of the previous state and closing a channel marks the completion of the communication on that channel. Here is 
an example showing how it fits together.

```go 
type Idler struct {
    next chan chan struct{}
}

func (i *Idler) AwaitIdle(ctx context.Context) error {
    idle := <-i.next
    i.next <- idle
    if idle != nil {
        select {
        case <-ctx.Done(): 
            return ctx.Err()
        case <-idle:
        }
    }
    return nil
}

func (i *Idler) SetBusy(b bool) {
    idle := <-i.next
    if b && (idle == nil) {
        idle = make(chan struct{})
    } else if !b && (idle != nil) {
        close(idle) // idle Now
        idle = nil
    }
    i.next <- idle
}

func NewIdler() *Idler {
    next := make(chan [...], 1)
    next <- nil
    return &Idler{next}
}
```

SetBusy allocates a new channel on the idle to busy transition and closes the previous channel if any on the busy to idle transition.

## Broadcast events
Broadcast may also signal ephemeral events such as configuration reload requests. 

### Events can be data
We can treat broadcast events like data updates and send them individually to each interested subscriber. The OS Signal package in 
the standard library takes that approach so that waiters can receive multiple events on the same channel. Alternatively we can treat
the event as the completion of the hasn't happened yet state and indicate it by closing a channel. That typically results in fewer
channel allocations but when we have closed the channel we can't communicate any additional data about the event.

**When we share by communicating, we should communicate the things that we want to share not just messages about them**.

# Worker Pools
We started with asynchronous patterns which deal with Goroutines, then we looked at condition variables which sometime deals with
resources. Now let's put them together.

The **worker pool** is a pattern that treats a set of Goroutines as resources. Just a note on terminology here, in other languages this
pattern is usually called a **Thread Pool**.

```go 
// start the workers
work := make(chan Task)
for n := limit; n > 0; n-- {
    go func() {
        for task := range work {
            perform(task)
        }
    }()
}

// Send the work
for _, task := range hugeSlice {
    work <- task
}
```

In the worker pool pattern we start up a fixed number of worker Goroutines that each read and perform tasks from a channel. another
Goroutine often the same one that started the workers sends the tasks to the workers, the sender blocks until the worker is available
to receive the next task. 

In laguages with heavyweight threads the **Worker Pool** pattern allows us to reuse threads for multiple tasks avoiding the overhead
of creating and destroying threads for small amounts of work. This benifit doesn't apply in Go because:

**Goroutines are multiplexed onto multiple OS threads. Their design hides many of the complexities of thread creation and management.**

The benefits that WorkerPool do provide in Go is to **limit the amount of concurrent work in flight**. If each task needs some 
limited resource, such as file handles, network bandwidth or even a non-trivial amount of RAM, a worker pool can bound the peak
resource usage of the program.

## Worker Lifetimes
The above worker pool has a problem it leakes the workers forever. If the API that we are implementing is synchronous and remember
what we said before about synchronous APIs. Or if we want to be able to reset the Worker State for unittest then we need to be able
to shutdown the workers and know when they have finished.

## Worker Pool: Cleaning up 
Cleaning up the workers adds a bit more boilerplate to the pattern. First we'll add a waitgroup to track the Goroutines.

```go 
// start the workers
work := make(chan Task)
var wg sync.WaitGroup 
for n := limit; n > 0; n-- {
    wg.Add(1)
    go func() {
        for task := range work {
            perform(task)
        }
        wg.Done()
    }()
}
```

The after we send the work we can close the channel and wait for the workers to exit.

```go 
//send the work
for _, task := range hugeSlice {
    work <- task
}

// Signal End of work
close(work)

// Wait for completion
wg.Wait()
```

## Idle workers
But we may have another problem, even if we remember to cleanup the workers when we're done we may leave them idle for the longtime.
Especially towards the end of the work may be forever if we've accidentally deadlocked something. Assuming we've remembered to cleanup
if we have deadlock our tests will hang instead of passing. So atleast we can get a Goroutine dump to help debug, but that's lot 
harder to find especially if our program happens to be a large service implemented with several different pools. It will also be a 
problem if we want to use the Goroutine dump to debug other issues such as crashes or memory-leaks.

Goroutines are lightweight but not free those idel workers still have a resource cost too and for large pool that cost may not be 
completely negligible.

So let's rethink this pattern. How can we get the same benefits as worker pools without the complexity of workers and their lifetimes?
We want to start the Goroutines only when we're actually ready to do the work and let them exit as soon as the worker is done.
Let's do just that part and see where we endup.

```go 
// Start the work
var wg sync.WaitGroup
for _, task := range hugeSlice {
    wg.Add(1)
    go func(task Task) {
        perform(task)
        wg.Done()
    }(task)
}

// Wait for completion
wg.Wait()
```

If we only need to distribute work across threads we can omit the workerpool and it's channel and use only the Waitgroup. This code 
is lot simpler but now we need to figure out how to limit the inflight work again.

We already have a pattern **Share resources by communicating the resources** for that. Remember limits are resources, so let's use
the semaphore channel pattern. The Semaphore example in effective go requires a token inside the Goroutine but we'll require it 
earlier right where we have the WaitGroup add call. We don't want a lot of Goroutines sitting around doing nothing and this way we 
have only one idle Goroutine instead of many.

```go 
// start the work
sem := make(chan token, limit)
for _, task := range hugeSlice {
    sem <- token{}
    go func(task Task) {
        perform(task)
        <-sem
    }(task)
}

// Wait for completion
for n := limit; n > 0; n-- {
    sem <- token{}
} 
```

Recall that we acquire the semaphore by sending a token and we release it by discarding a token.

Now the semaphore fits in pretty nicely in placee of the waitgroup and that's no accident. ```sync.WaitGroup``` is very similar to 
a semaphore, the only major difference is that the WaitGroup allows further ```Add``` calls during ```Wait```. Whereas our wait loop
on our semaphore channel doesnot. Fortunately that usually doesn't matter as in this case.

Remember our first worker pool with the two loops and how we leaked all thos idle workers forever, if you look carefully these (above 
one) are the same two loops swapped around:

```go 
// start the workers
work := make(chan Task)
for n := limit; n > 0; n-- {
    go func() {
        for task := range work {
            perform(task)
        }
    }()
}

// Send the work
for _, task := range hugeSlice {
    work <- task
}
```

We have eliminated the leak without adding any net lines of code.

# Recap
- Start Goroutines when you have concurrent work to do immediately, don't contort the API to avoid blocking the caller and don't 
spoolup idle workers that will just fillup your Goroutine dumps.. It's easy to start Goroutines when you need them and it's easy
to block when you don't.
- Share things by communicating those things directly. Opaque signals about shared memory make it entirely too easy to send the 
signals to the wrong place or miss them entirely. Instead communicate where things need to go and then communicate to send them 
there.
