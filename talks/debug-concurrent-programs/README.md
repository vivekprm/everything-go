Concurrent programs doesn't always have reproducible behavior. E.g.

```go
package main

import (
    "fmt"
    "time"
)

func say(s string) {
    for i := 0; i < 5; i++ {
        time.Sleep(100 * time.Millisecond)
        fmt.Println(s)
    }
}

func main() {
    go say("world")
    say("hello")
}
```

Everytime we run above program we see different output.

# What is a Goroutine?
Goroutine is just like an abstraction. Below is the struct that handles the Goroutine under the hood.

```go
type g struct {
	// Stack parameters.
	// stack describes the actual stack memory: [stack.lo, stack.hi).
	// stackguard0 is the stack pointer compared in the Go stack growth prologue.
	// It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
	// stackguard1 is the stack pointer compared in the //go:systemstack stack growth prologue.
	// It is stack.lo+StackGuard on g0 and gsignal stacks.
	// It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	stack       stack   // offset known to runtime/cgo
	stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink

	_panic    *_panic // innermost panic - offset known to liblink
	_defer    *_defer // innermost defer
	m         *m      // current m; offset known to arm liblink
	sched     gobuf
	syscallsp uintptr // if status==Gsyscall, syscallsp = sched.sp to use during gc
	syscallpc uintptr // if status==Gsyscall, syscallpc = sched.pc to use during gc
	syscallbp uintptr // if status==Gsyscall, syscallbp = sched.bp to use in fpTraceback
	stktopsp  uintptr // expected sp at top of stack, to check in traceback
	// param is a generic pointer parameter field used to pass
	// values in particular contexts where other storage for the
	// parameter would be difficult to find. It is currently used
	// in four ways:
	// 1. When a channel operation wakes up a blocked goroutine, it sets param to
	//    point to the sudog of the completed blocking operation.
	// 2. By gcAssistAlloc1 to signal back to its caller that the goroutine completed
	//    the GC cycle. It is unsafe to do so in any other way, because the goroutine's
	//    stack may have moved in the meantime.
	// 3. By debugCallWrap to pass parameters to a new goroutine because allocating a
	//    closure in the runtime is forbidden.
	// 4. When a panic is recovered and control returns to the respective frame,
	//    param may point to a savedOpenDeferState.
	param        unsafe.Pointer
	atomicstatus atomic.Uint32
	stackLock    uint32 // sigprof/scang lock; TODO: fold in to atomicstatus
	goid         uint64
	schedlink    guintptr
	waitsince    int64      // approx time when the g become blocked
	waitreason   waitReason // if status==Gwaiting

	preempt       bool // preemption signal, duplicates stackguard0 = stackpreempt
	preemptStop   bool // transition to _Gpreempted on preemption; otherwise, just deschedule
	preemptShrink bool // shrink stack at synchronous safe point

	// asyncSafePoint is set if g is stopped at an asynchronous
	// safe point. This means there are frames on the stack
	// without precise pointer information.
	asyncSafePoint bool

	paniconfault bool // panic (instead of crash) on unexpected fault address
	gcscandone   bool // g has scanned stack; protected by _Gscan bit in status
	throwsplit   bool // must not split stack
	// activeStackChans indicates that there are unlocked channels
	// pointing into this goroutine's stack. If true, stack
	// copying needs to acquire channel locks to protect these
	// areas of the stack.
	activeStackChans bool
	// parkingOnChan indicates that the goroutine is about to
	// park on a chansend or chanrecv. Used to signal an unsafe point
	// for stack shrinking.
	parkingOnChan atomic.Bool
	// inMarkAssist indicates whether the goroutine is in mark assist.
	// Used by the execution tracer.
	inMarkAssist bool
	coroexit     bool // argument to coroswitch_m

	raceignore    int8  // ignore race detection events
	nocgocallback bool  // whether disable callback from C
	tracking      bool  // whether we're tracking this G for sched latency statistics
	trackingSeq   uint8 // used to decide whether to track this G
	trackingStamp int64 // timestamp of when the G last started being tracked
	runnableTime  int64 // the amount of time spent runnable, cleared when running, only used when tracking
	lockedm       muintptr
	fipsIndicator uint8
	sig           uint32
	writebuf      []byte
	sigcode0      uintptr
	sigcode1      uintptr
	sigpc         uintptr
	parentGoid    uint64          // goid of goroutine that created this goroutine
	gopc          uintptr         // pc of go statement that created this goroutine
	ancestors     *[]ancestorInfo // ancestor information goroutine(s) that created this goroutine (only used if debug.tracebackancestors)
	startpc       uintptr         // pc of goroutine function
	racectx       uintptr
	waiting       *sudog         // sudog structures this g is waiting on (that have a valid elem ptr); in lock order
	cgoCtxt       []uintptr      // cgo traceback context
	labels        unsafe.Pointer // profiler labels
	timer         *timer         // cached timer for time.Sleep
	sleepWhen     int64          // when to sleep until
	selectDone    atomic.Uint32  // are we participating in a select and did someone win the race?

	// goroutineProfiled indicates the status of this goroutine's stack for the
	// current in-progress goroutine profile
	goroutineProfiled goroutineProfileStateHolder

	coroarg   *coro // argument during coroutine transfers
	syncGroup *synctestGroup

	// Per-G tracer state.
	trace gTraceState

	// Per-G GC state

	// gcAssistBytes is this G's GC assist credit in terms of
	// bytes allocated. If this is positive, then the G has credit
	// to allocate gcAssistBytes bytes without assisting. If this
	// is negative, then the G must correct this by performing
	// scan work. We track this in bytes to make it fast to update
	// and check for debt in the malloc hot path. The assist ratio
	// determines how this corresponds to scan work debt.
	gcAssistBytes int64
}
```

It's defined [here](https://go.dev/src/runtime/runtime2.go)

Different Goroutines are multiplexed over different OS threads. 

It hides lots of design and complexity of thread creation and management.

# How Can I debug my concurrent program?
- Visualize Goroutine
    - Use colored logs using [this package](https://github.com/xiegeo/coloredgoroutine).
    - https://divan.dev/posts/go_concurrency_visualize/
    - Print how go schedule events using env var: ```GODEBUG=schedtrace=5000 <binary>```
    - Using Debuggers e.g. delve, gdb
```go
package main

import (
	"fmt"
	"os"
	"sync"
	"github.com/xiegeo/coloredgoroutine"
	"github.com/xiegeo/coloredgoroutine/goid"
)
func main() {
	c := coloredgoroutine.Colors(os.Stdout)
	fmt.Fprintln(c, "Hi, I am a goroutine", goid.ID(), "from main routine")

	count := 10

	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		n := i
		go func() {
			fmt.Fprintln(c, "Hi, I am a goroutine", goid.ID(), "from loop i = ", n)
			wg.Done()
		}()
	}
	wg.Wait()
}
```

# How to set breakpoint inside Goroutine
Yes we can set breakpoints inside goroutine and also in channels. We get more details using dlv commandline.

However we can't send anything on the channel using dlv prompt. Open [issue](https://github.com/go-delve/delve/issues/2117)

# Debugging Goroutine
```
(dlv)Goroutine
```

Prints information about the current goroutine

# Debugging Goroutine
```
(dlv)Goroutine
```

Prints information about the current goroutine..

# Profile Labels
```go
labels := pprof.Labels("worker", "purge")
pprof.Do(ctx, labels, func(ctx context.Context) {
    // Do some work...

    go update(ctx)  // propagates labels in ctx
})
```

You can use profile labels inside pprof model, it marks your goroutines with labels and usually you use these lables for profiling so you can open pprof profiles and see like some different metrics. But you can also do it with delve which is supercool.

If you label your goroutines with labels:

```go
go func (p string, rid int64) {
    labels := pprof.Labels("request", "automated", "page", p, "rid", str)
    pprof.Do(context.Background(), labels, func(_ context.Context) {
        makeRequest(activeConns, c, p, rid)
    })
}(page, i)
```

Or you can use debugger middleware using [this library](https://github.com/dlsniper/debugger), which add labels to all your handlers, which is nice:

```go
router.HandleFunc("/", debugger.Middleware(homeHandler, func(r *http.Request) []string {
    return []string {
        "path", r.RequestURI
    }
}))
```

You can also add it directly using ```SetLabels``` using this library:

Original
```go
func sum(a, b int) int {
    return a + b
}
```

Replacement:
```go
func sum(a, b int) int {
    debugger.SetLabels(func() []string {
        return []string{
            "a", strconv.Itoa(a),
            "b", strconv.Itoa(b),
        }
    })
    return a + b
}
```

Then if you run goroutines inside delve debugger:
```sh
(dlv)goroutines -l
```

It will print goroutines without labels. But if you run below

```sh
(dlv)goroutines -l -with label page=/about
```

Will print gorutines for provided label.

You can play on your own using [this](https://github.com/dlsniper/serverdemo) example project.

```sh
dlv debug --build-flags="-ldflags=-s -tags=debugger" *.go
```

# Experiment
You can play on your own using example project:

https://github.com/dlsniper/serverdemo

```sh
dlv debug --build-flags="-ldflags=-s -tags=debugger" *.go
```

# GDB & Golang
```sh
go build -ldflags=-compressdwarf=false -gcflags=all="-N -I" -o main main.go
```

It has not many features:

```sh
(gdb) info goroutines
(gdb) bt
(gdb) goroutine 1 bt
```

You can't filter goroutines and it's not very readable. So it's advised to use delve.

# Deadlocks happen and are painful to debug
One important problem in Golang concurrency world is deadlocks, with deadlocks usually program get stuck on the channel send operation which is waiting forever for example to read the value.

```go
package main

func main() {
	ch := make(chan string)
	ch <- "hello deadlock"
}
```

Golang supports detection of these situations.

Another example:
```go
func main(){
	c := make(chan bool)
	m := make(map[string]string)
	go func() {
		m["1"] = "a" // First conflicting access
		c <- true
	}()
	m["2"] = "b" // Second conflicting access
	<- c
	for k, v := range m {
		fmt.Println(k, v)
	}
}
```

```sh
go run -race race_demo.go
```

```-race``` can't always find all data races.

## Real world examples complicated scenario & tools
https://github.com/sasha-s/go-deadlock

Using this library lots of deadlocks were found on cockroachdb and lots of interesting example of how mutex can be handled properly, how to write it properly etc. 

# 7 Simple rules for debugging concurrent applications
- Never assume a particular order of execution.
- Implement concurrency at the highest level possible.
- Don't forget Go only detects when the program as a whole freezes, not when a subset of Goroutines get stuck.
- STRACE: helps us to see are we waiting for some resource like reading file, access network etc.
- Conditional breakpoints to help you cover cases especially when it's concurrent program so you can catch only your case not like click next on every Goroutine.
- As discussed use Shadowing tracer 
  - ```DEBUG=schdtrace=5000```
- go-deadlock
- Use delve debugger

# References
- https://github.com/golang/go/blob/release-branch.go1.14/src/runtime/HACKING.md
- https://github.com/golang/go/wiki/LearnConcurrency
- https://rakyll.org/go-cloud
- https://yourbasic.org/golang/detect-deadlock
- https://blog.minio.io/debugging-goroutine-leaks-a1220142d32c
- https://golang.org/src/cmd/link/internal/ld/dwarf.go
- https://golang.org/src/runtime/runtime-gdb.py
- https://cseweb.ucsd.edu/~yiying/GoStudy-ASPLOS19.pdf
- https://golang.org/doc/articles/race_detector.html
- https://go.dev/doc/articles/race_detector