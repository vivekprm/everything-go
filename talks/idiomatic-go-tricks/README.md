https://youtu.be/yeetIgNeIkc?si=8HsZA3gmQWKN6znk

Idiomatic is an adjective meaning: using, containing or denoting expressions that are natural to native speakers.


# GO Code 
```go
func BrilliantFunction() {
    something, err := GetSomething()
    if err != nil {
        return nil, err
    }
    defer something.Close()
    if !something.OK() {
        return nil, errors.New("something went wrong")
    }
    another, err := something.Else()
    if err != nil {
        return nil, &customErr{err: err, location: "BrilliantFunction"}
    }
    another.Lock()
    defer another.Unlock()
    err := another.Update(1)
    if err != nil {
        return nil, err
    }
    return another.Thing(), nil
}
```

This is what most Go code ends up looking like.

- We don't have lots of line space gaps between the GO code and it's all kinds of sits together neatly. That's also one of the features. 
- We also have multiple return arguments, last return argument is usually an error.
- Defer is used because it expresses very clearly what you are doing and you expect something to close regardless in this case.

# Line of Sight
- Definition: a stright line along which an observer has unobstructed vision.
- Happy path is aligned to the left.
- Quickly scan a function to see expected flow.
- Error handling and edge cases indented.

# Bad Line of Sight
```go
func UnbrilliantFunction() (*Thing, error) {
    something, err := GetSomething()
    if err != nil {
        return nil, err
    }
    defer something.Close()
    if something.OK() {
        another, err := something.Else()
        if err != nil {
            return nil, &customErr{err: err, location: "BrilliantFunction"}
        }
        another.Lock()
        err = another.Update(1)
        if err == nil {
            another.Unlock()
            return another.Thing(), nil
        }
        another.Unlock()
        return nil, err
    } else {
        return nil, errors.New("something went wrong")
    }
}
```

This time it's difficult to see what's going on as we have more nesting in this case.

# Line of Sight Tips
- Make happy return that last statement if possible.
- Next time you write else, consider flipping the logic.

```go
if something.OK() {
    // do stuff
    return true, nil
} else {
    return false, errors.New("something")
}
```

becomes:

```go
if !something.OK() {
    return false, errors.New("something") 
}
return true, nil
```

See also **Cyclomatic complexity** - the persuit of simplicity.

# Single Method Interfaces
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

- Interface consisting on only one method
- Simpler = more powerful and useful
- Easy to implement
- Used throughout the standard library

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

- Only need to implement one method on a struct to build a handler.

You only need to build a struct with one method it's awesome but you can go a bit further than that, it's another classic Go thing which feels little bit weired initially but once you get your head around it. It turns out to be extremely useful.

# Function type alternatives for Single Method Interfaces
```http.Handler``` has a counterpart called ```http.HanderFunc```

```go
// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// [Handler] that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

It's a function type with a method that implements the Handler interface. So it's a function that has it's own method on it and when you call that method it just calls itself, it just calls the function but what that means is, you now no longer need the struct at all you can just do a function and you can use this pattern yourself, if you discover single method interfaces consider whether there's a use for just having a Func kind of alternative.

# Log Blocks
When we have lots of logs it's hard to find what we are looking at and it's hard to see the function we are working on.

```Go
func foo() error {
    log.Println("--------------")
    defer log.Println("----------------")

    // ...
}
```

It nicely tells us the log boundary of the function we are working on.
Easy to comment and remove.

# Return Teardown Functions
You have some function that does some work, great example is in testing, where you have some setup code, like here we have a setup function it's going to create a sample file for us to use and really that file needs to be cleaned up after we need to close it and also delete it.

```go
func setup(t *testing.T) (*os.File, func, error) {
    teardown := func() {}
    // make a test file
    f, err := ioutil.TempFile(os.TempDir(), "test")
    if err != nil {
        return nil, teardown, err
    }
    teardown = func() {
        // Close f
        err := f.Close()
        if err != nil {
            t.Error("setup: Close:", err)
        }
        // delete the test file
        err = os.RemoveAll(f.Name())
        if err != nil {
            t.Error("setup: RemoveAll:", err)
        }
    }
    return f, teardown, nil
}
```

So what we can do when we call it, we can immediately defer the teardown:

```go
func TestSomething(t *testing.T) {
    f, teardown, err := setup(t)
    defer teardown()
    if err != nil {
        t.Err("setup:", err)
    }
    // do something with f
}
```

- Cleanup code is encapsulated.
- Caller doesn't need to worry about cleaning up.
- If setup changes, code that uses it doesn't necessarily need to.

# Good Timing
Another way to use kind of returning functions is in timing. We use this exact pattern:

```go
func StartTimer(name string) func() {
    t := time.Now()
    log.Println(name, "started")
    return func() {
        d := time.Now().Sub(t)
        log.Println(name, "took", d)
    }
}

func FunkyFunc() {
    stop := StartTimer("FlunkyFunc")
    defer stop()

    timer.Sleep(1 * time.Second)
}
```

- Capture the state in the Closure.
- Make things easy for your users.

Calling code doesn't have to know about the time or doesn't have to worry about the state. It's all kind of captured in that Closure, so client doens't have to trouble itself with it.


# Discover Interfaces
Here we have Sizer interface and idea is we want to get size of something and so we have single method interface, which is nice. We have couple of functions where we could use that.

```go
// Sizer describes the Size() method that gets the 
// total size of an item
type Sizer interface {
    Size() int64
}

func Fits(capacity int64, v Sizer) bool {
    return capacity > v.Size()
}

func IsEmailable(v Sizer) bool {
    return 1 << 20 > v.Size()
}

// Size gets the size of a File
func (f *File) Size() int64 {
    return f.info.Size()
}
```

It's quite easy to implement this interface.

## Many items as one
Then we can do something quite clever which is, you can create a new type which is a slice of that interface and make that type implement the same interface.

```go
type Sizers []Sizer

func (s Sizers) Size() int64 {
    var total int64

    for _, sizer := range s {
        total += sizer.Size()
    }
    return total
}
```

- The slice type implements the Sizer interface.
- Now a set of objects can be used in **Fits** and **IsEmailable** functions.

Now you can treat many items as one and you can use them in the same place as you use the Size(). So it goes into the Fits method not problem. So we could have 10 files and pass that in and Fits method doesn't have to know that's what you are doing. The impelemtation here just iterates over the objects, calls Size and totals it up.


# Otherways to Implement the Interface
```go
type Sizer interface {
    Size() int64
}

type SizeFunc func() int64

func (s Sizefunc) Size() int64 {
    return s()
}

type Size int64

func (s Size) Size() int64 {
    return int64(s)
}
```

- SizeFunc means we can write ad-hoc size calculator and still make use of the same methods.
    - We have a function that ad-hoc calculates the size of something and we can pass that into anything that is using this Sizer interface.
- Size int64 type means we can specify explicit sizes: Size(123)
    - This type implements the Size interface it just returns itself. So now I can just pass wherever I need to pass as Sizer I don't need to create an object. I can just pass in I can just cast it.
- easy because interface is so small.

# Optional Features
Really quite powerful.

```go
type Valid interface {
    OK() error
}

func (p Person) OK() error {
    if p.Name == "" {
        return errors.New("name required")
    }
    return nil
}
func Decode(r io.Reader, v interface{}) error {
    err := json.NewDecoder(r).Decode(v)
    if err != nil {
        return err
    }
    obj, ok := v.(Valid)
    if !ok {
        return nil // no OK method
    }
    err = obj.OK()
    if err != nil {
        return err
    }
    return nil
}
```

Look at the Decode function and the idea is we are passing a Reader, probably something from a HTTP request and it's going to decode the JSON for us. But optionally it's also going to call this Valid interface method called OK. It's going to call that on the object if it has it and the way we do that here is by casting and doing type assertion.

Keeping the Validation of the Object with the Object rather than repeating it or having it in some other place. You can do that for many different things.

So wehnever you think you know, we have this kind of thing and sometimes it's not quite always like this, sometimes it also has these extra bits and pieces, so then consider may be there's two interfaces or maybe it's worth abstracting.

# Simple Mocks
Here we have an interface called MailSender and it has two methods. This is nice little trick of creating something that you can use in testing. So you just make a struct that has fields for each method. Because you know functions just a type really. So these functions SendFunc and SendFromFunc can be set when you are writing tests and then all we do is implement the interface that just calls those functions.


```go
type MailSender interface {
    Send(to, subject, body string) error
    SendFrom(from, to, subject, body string) error
}

type MockSender struct {
    SendFunc        func(to, subject, body string) error
    SendFromFunc    func(from, to, subject, body string) error
}

func (m MockSender) Send(to, subject, body string) error {
    return m.SendFunc(to, subject, body)
}

func (m MockSender) SendFrom(from, to, subject, body string) error {
    return m.SendFromFunc(from, to, subject, body)
}
```

So you don't have to build complicated mock objects what we can do in test code is just create this mock version and pass it into our code that we're testing and we can control in the test what that function is going to do. So that could involve capturing what was called and making assertions about it and you can also control what is returned. So for example in case of SendFromFunc we are returning error.

Another nice thing is you don't have to implement every field you can focus on the one you want to test if other fields are called you're going to get some kind of nil Panic probably.

```go
func TestWelcomeEmail(t *testing.T) {
    errTest := errors.New("nope")
    var msg string

    sender := MockSender{
        SendFunc: func(to, subject, body string) error {
            msg = fmt.Sprintf("(%s) %s: %s", to, subject, body)
            return nil
        },
        SendFromFunc: func(from, to , subject, body string) error {
            return errTest
        }
    }

    SendWelcomeEmail(sender, "to", "subject", "body")

    if msg != "(to) subject: body" {
        t.Error("SendWelcomeEmail:", msg)
    }
}
```

# Mocking Other people's struct
Sometimes somebody else provides the struct (and not an interface) and you wish it was an interface.

```go
package them

type Messenger struct {}

func (m *Messenger) Send(to, message string) error {
    // ...their code
}
```

But don't despair make your own interface.

```go
type Messenger interface {
    Send(to, message string) error
}
```

- This interface is already implemented by ```them.Messenger``` struct.
- We are free to mock it in test code or even provide our own implementation.


# Retrying
```go
type Func func(attempt int) (retry bool, err error)

func Try(fn Func) error {
    var err error
    var cont bool
    attempt := 1
    for {
        cont, err := fn(attempt)
        if !cont || err == nil {
            break
        }
        attempt++
        if attempt > MaxRetries {
            return errMaxRetriesReached
        }
    }
    return err
}
```

Full code at https://github.com/matryer/try

Retrying 5 times:

```go
var value string
err := Try(func(attempt int) (bool, error) {
    var err error
    value, err := SomeFunction()
    return attempt < 5, err // try 5 times
})

if err != nil {
    log.Fatalln("error:", err)
}
```

- Return whether to retry or not, or an error
- Easy to read

## Retrying: Delay between retries

```go
var value string

err := Try(func(attempt int) (bool, error) {
    var err error
    value, err := SomeFunction()
    if err != nil {
        time.Sleep(1 * time.Minute) // wait a minute
    }
    return attempt < 5, err
})

if err != nil {
    log.Fatalln("error:", err)
}
```

- Don't bloat the Try function with arguments.
- Let people do extra things with Go code where possible.

# Empty Struct Implementations
Sometimes we have interface like ```Codec``` and you want to implement that Codec. We can do it like below with an empty struct that essentially is just a collection of these the methods that implement that interface. Since there is no state in here there's no need for it to be anything more complicated. 

```go
try Codec interface {
    Encode(w io.Writer, v interface{}) error
    Decode(r io.Reader, v interface{}) error
}

type jsonCodec struct{}
func (jsonCodec) Encode(w io.Writer, v interface{}) error {
    return json.NewEncoder(w).Encode(v)
}
func (jsonCoded) Decode(r io.Reader, v interface{}) error {
    return json.NewDecoder(r).Decode(v)
}

var JSON codec = jsonCodec{}
```

You can see that the receiver the jsonCodec, the receiver actually we don't capture it. It just says the type and that also makes it clear that we're not going to use or we're not going to store any state.

What's nice about this is, if this was a package all we have to expose is the ```Codec``` interface and the JSON variable. So it's dead clear that what's going on. We don't have to have ```jsonCodec``` expose as well. That's why it starts with lowercase.

Other nice thing is, wherever you need if you want to just use the json stuff directly if it was in encoding package ```encoding.JSON.Encode()``` and access directly those methods without exposing the type.

# Semaphores
Limit the number of ```Goroutines``` running at once.

```go
var (
    concurrent = 5
    semaphoreChan = make(chan struct{}, concurrent)
)

func doWork(item int) {
    semaphoreChan <- struct{}{} // block while full
    go func() {
        defer func() {
            <- semaphoreChan    // read to release a slot
        }()
        log.Println(item)
        time.Sleep(1 * time.Second)
    }()
}

func main() {
    for i := 0; i < 10000; i++ {
        doWork(i)
    }
}
```

Essentially the idea is you create aa bufferred channel and size of the channel is how many concurrent Goroutines you want and what you do before you do the work, you send in something into that channel. And while there is room in that channel that's going to go through no problem. As soon as that channel is full it'll block it can't put anymore in. 

You spin up another Goroutine here where you defer the function of reading from that semaphore channel and that then clears one of the slots which would then unblock one of the others.

# Be Obvious not Clever
```go
func something() {
    defer StartTimer("something")()

    // :-|
}
```

Should be:

```go
func something() {
    stop := StartTimer()
    defer stop()

    // :-)
}
```

# How to become a native speaker
- Read the standard library
- Write the obvious code (not clever)
- Don't surprise your users
- Seek simplicity

Learn from others:
- Participate in open-source projects
- Ask for reviews and accept criticisms
- Help others when you spot something (and be kind)

Follow @matryer


