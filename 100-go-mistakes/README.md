# 1: Unintended variable shadowing
In Go, a variable name declared in a block can be redeclared in an inner block. This principle, called **variable shadowing**, is prone to common mistakes.

The following example shows an unintended side effect because of a shadowed variable. It creates an HTTP client in two different ways, depending on the value of a tracing Boolean:

```go
var client *http.Client // Declares a client variable
if tracing {
    // Creates a HTTP client with tracing enabled. (The client variable is shadowed 
    // in this block)
    client, err := createClientWithTracing()
    if err != nil {
        return err
    }
    log.Println(client)
} else {
    // Creates a default HTTP client. (The client variable is also shadowed 
    // in this block)
    client, err := createDefaultClient()
    if err != nil {
        return err
    }
    log.Println(client)
}
// Use client
```

In this example, we first declare a client variable. Then, we use the short variable declaration operator (:=) in both inner blocks to assign the result of the function call to the inner client variables—not the outer one. As a result, the outer variable is always nil.

**NOTE** This code compiles because the inner client variables are used in the logging calls. If not, we would have compilation errors such as client declared and not used.

We can fix it as below:

```go
var client *http.Client
if tracing {
    client, err = createClientWithTracing()
} else {
    client, err = createDefaultClient()
}
if err != nil {
    // Common error handling
}
// use client
```

We should remain cautious because we now know that we can face a scenario where the code compiles, but the variable that receives the value is not the one expected.

# 2: Unnecessary nested code
Code is qualified as readable based on multiple criteria such as naming, consistency, formatting, and so forth.
Readable code requires less cognitive effort to maintain a mental model; hence, it is easier to read and maintain.

A critical aspect of readability is the number of nested levels. Let’s do an exercise.

Suppose that we are working on a new project and need to understand what the following join function does:
```go
func join(s1, s2 string, max int) (string, error) {
    if s1 == "" {
        return "", errors.New("s1 is empty")
    } else {
        if s2 == "" {
            return "", errors.New("s2 is empty")
        } else {
            // Calls a concatenate function to perform some specific concatenation
            // but may return errors
            concat, err := concatenate(s1, s2)
            if err != nil {
                return "", err
            } else {
                if len(concat) > max {
                    return concat[:max], nil
                } else {
                    return concat, nil
                }
            }
        }
    }
}
func concatenate(s1 string, s2 string) (string, error) {
    // ...
}
```

This join function concatenates two strings and returns a substring if the length is greater than max. Meanwhile, it handles checks on s1 and s2 and whether the call to concatenate returns an error.

From an implementation perspective, this function is correct. However, building a mental model encompassing all the different cases is probably not a straightforward task. Why? Because of the number of nested levels.

Now, let’s try this exercise again with the same function but implemented differently:
```go
func join(s1, s2 string, max int) (string, error) {
    if s1 == "" {
        return "", errors.New("s1 is empty")
    }
    if s2 == "" {
        return "", errors.New("s2 is empty")
    }
    concat, err := concatenate(s1, s2)
    if err != nil {
        return "", err
    }
    if len(concat) > max {
        return concat[:max], nil
    }
    return concat, nil
}
func concatenate(s1 string, s2 string) (string, error) {
    // ...
}
```

You probably noticed that building a mental model of this new version requires less cognitive load despite doing the same job as before. Here we maintain only two nested levels. 
As mentioned by Mat Ryer, a panelist on the Go Time podcast (https://
medium.com/@matryer/line-of-sight-in-code-186dd7cdea88): 

*Align the happy path to the left; you should quickly be able to scan down one column to see the expected execution flow*.

It was difficult to distinguish the expected execution flow in the first version because of the nested if/else statements. Conversely, the second version requires scanning down one column to see the expected execution flow and down the second column to see how the edge cases are handled.

When an if block returns, we should omit the else block in all cases. For example, we shouldn’t write
```go
if foo() {
    // ...
    return true
} else {
    // ...
}
```

Instead, we omit the else block like this:
```go
if foo() {
    // ...
    return true
}
// ...
```

Writing readable code is an important challenge for every developer. Striving to reduce the number of nested blocks, aligning the happy path on the left, and returning as early as possible are concrete means to improve our code’s readability.

# 3: Misusing init functions
Sometimes we misuse init functions in Go applications. The potential consequences are poor error management or a code flow that is harder to understand. Let’s refresh our minds about what an init function is. Then, we will see when its usage is or isn’t recommended.

## Concepts
An init function is a function used to initialize the state of an application. It takes no arguments and returns no result (a func() function). **When a package is initialized, all the constant and variable declarations in the package are evaluated. Then, the init functions are executed**. Here is an example of initializing a main package:

```go
package main

import "fmt"

// Executed first
var a = func() int {
    fmt.Println("var")
    return 0
}()

// Executed second
func init() {
    fmt.Println("init")
}

// Executed last
func main() {
    fmt.Println("main")
}
```

Running this example prints the following output:
```
var
init
main
```

An init function is executed when a package is initialized. In the following example, we define two packages, ```main``` and ```redis```, where ```main``` depends on ```redis```. First, ```main.go``` from the ```main``` package:

```go
package main

import (
    "fmt"
    "redis"
)

func init() {
    // ...
}

func main() {
    // A dependency on the redis package
    err := redis.Store("foo", "bar")
    // ...
}
```

And then ```redis.go``` from the redis package:
```go
package redis

// imports

func init() {
    // ...
}
func Store(key, value string) error {
    // ...
}
```

Because ```main``` depends on ```redis```, the ```redis``` package’s ```init``` function is executed first, followed by the ```init``` of the ```main``` package, and then the ```main``` function itself. Figure 2.2 shows this sequence.

We can define multiple ```init``` functions per package. When we do, the execution order of the ```init``` function inside the package is based on the source files’ alphabetical order. For example, if a package contains an ```a.go``` file and a ```b.go``` file and both have an init function, the ```a.go``` init function is executed first.

We shouldn’t rely on the ordering of ```init``` functions within a package. Indeed, it can be
dangerous as source files can be renamed, potentially impacting the execution order.

We can also define multiple ```init``` functions within the same source file. For exam-
ple, this code is perfectly valid:
```go
package main

import "fmt"

// First init function
func init() {
    fmt.Println("init 1")
}

// Second init function
func init() {
    fmt.Println("init 2")
}

func main() {
}
```

The first init function executed is the first one in the source order. Here’s the output:
```
init 1
init 2
```

We can also use ```init``` functions for side effects. In the next example, we define a main package that doesn’t have a strong dependency on ```foo``` (for example, there’s no direct
use of a public function). However, the example requires the ```foo``` package to be initialized. We can do that by using the _ operator this way:

```go
package main
import (
"fmt"
_ "foo" // Imports foo for side effects
)
func main() {
    // ...
}
```

In this case, the ```foo``` package is initialized before main. Hence, the init functions of foo
are executed.
Another aspect of an ```init``` function is that it can’t be invoked directly, as in the following example:

```go
package main
func init() {}

func main() {
    init()  // Invalid reference
}
```

This code produces the following compilation error:
```
$ go build .
./main.go:6:2: undefined: init
```

Now that we’ve refreshed our minds about how ```init``` functions work, let’s see when we
should use or not use them. The following section sheds some light on this.

## When to use init functions
First, let’s look at an example where using an ```init``` function can be considered inappropriate: holding a database connection pool. In the ```init``` function in the example, we
open a database using ```sql.Open```. We make this database a global variable that other
functions can later use:

```go
var db *sql.DB

func init() {
    dataSourceName := os.Getenv("MYSQL_DATA_SOURCE_NAME") // Environment variable
    d, err := sql.Open("mysql", dataSourceName)
    if err != nil {
        log.Panic(err)
    }
    err = d.Ping()
    if err != nil {
        log.Panic(err)
    }
    db = d  // Assigns the DB connection to the global db variable
}
```

In this example, we open the database, check whether we can ping it, and then assign it to the global variable. What should we think about this implementation? Let’s describe three main downsides.

- First, error management in an ```init``` function is limited. Indeed, as an ```init``` function doesn’t return an error, one of the only ways to signal an error is to panic, leading the application to be stopped. In our example, it might be OK to stop the application anyway if opening the database fails. However, it shouldn’t necessarily be up to the package itself to decide whether to stop the application. 

Perhaps a caller might have preferred implementing a retry or using a fallback mechanism. In this case, **opening the database within an init function prevents client packages from implementing their error-handling logic**.

- Another important downside is related to testing. If we add tests to this file, **the init function will be executed before running the test cases**, which isn’t necessarily what we want (for example, if we add unit tests on a utility function that doesn’t require this connection to be created). Therefore, the init function in this example complicates writing unit tests.
- The last downside is that the example requires assigning the database connection pool to a global variable. Global variables have some severe drawbacks; for example:
  - Any functions can alter global variables within the package.
  - Unit tests can be more complicated because a function that depends on a global variable won’t be isolated anymore.

In most cases, we should favor encapsulating a variable rather than keeping it global.

For these reasons, the previous initialization should probably be handled as part of a plain old function like so:

```go
// Accepts a data source name and returns an *sql.DB and an error
func createClient(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err // // Returns an error
    }
    if err = db.Ping(); err != nil {
        return nil, err 
    }
    return db, nil
}
```

Using this function, we tackled the main downsides discussed previously. Here’s how:
- The responsibility of error handling is left up to the caller.
- It’s possible to create an integration test to check that this function works.
- The connection pool is encapsulated within the function.

Is it necessary to avoid init functions at all costs? Not really. There are still use cases
where init functions can be helpful. For example, the official Go blog (http://mng.bz/PW6w) uses an init function to set up the static HTTP configuration:

```go
func init() {
    redirect := func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/", http.StatusFound)
    }
    http.HandleFunc("/blog", redirect)
    http.HandleFunc("/blog/", redirect)
    static := http.FileServer(http.Dir("static"))
    http.Handle("/favicon.ico", static)
    http.Handle("/fonts.css", static)
    http.Handle("/fonts/", static)
    http.Handle("/lib/godoc/", http.StripPrefix("/lib/godoc/", http.HandlerFunc(staticHandler)))
}
```

In this example, the init function cannot fail (```http.HandleFunc``` can panic, but only if the handler is nil, which isn’t the case here). Meanwhile, there’s no need to create any global variables, and the function will not impact possible unit tests. Therefore, this code snippet provides a good example of where init functions can be helpful. In summary, we saw that init functions can lead to some issues:
- They can limit error management.
- They can complicate how to implement tests (for example, an external dependency must be setup, which may not be necessary for the scope of unit tests).
- If the initialization requires us to set a state, that has to be done through global variables.

We should be cautious with init functions. They can be helpful in some situations, however, such as defining static configuration, as we saw in this section. Otherwise, and in most cases, we should handle initializations through ad hoc functions.

# 4: Overusing getters and setters
In programming, data encapsulation refers to hiding the values or state of an object. Getters and setters are means to enable encapsulation by providing exported methods on top of unexported object fields.
In Go, there is no automatic support for getters and setters as we see in some languages. It is also considered neither mandatory nor idiomatic to use getters and setters to access struct fields. For example, the standard library implements structs in which some fields are accessible directly, such as the ```time.Timer``` struct:

```go
timer := time.NewTimer(time.Second)
<-timer.C   // C is a <–chan Time field
```

Although it’s not recommended, we could even modify C directly (but we wouldn’t receive events anymore). However, this example illustrates that the **standard Go library doesn’t enforce using getters and/or setters even when we shouldn’t modify a field**.

On the other hand, using getters and setters presents some advantages, including these:
- They encapsulate a behavior associated with getting or setting a field, allowing new functionality to be added later (for example, validating a field, returning a computed value, or wrapping the access to a field around a mutex).
- They hide the internal representation, giving us more flexibility in what we expose.
- They provide a debugging interception point for when the property changes at run time, making debugging easier.

If we fall into these cases or foresee a possible use case while guaranteeing forward compatibility, using getters and setters can bring some value. For example, if we use them with a field called balance, we should follow these naming conventions:
- The getter method should be named ```Balance``` (not ```GetBalance```).
- The setter method should be named ```SetBalance```

Here’s an example:
```go
currentBalance := customer.Balance()    // Getter
if currentBalance < 0 {
    customer.SetBalance(0)  // Setter
}
```

In summary, we shouldn’t overwhelm our code with getters and setters on structs if they don’t bring any value. We should be pragmatic and strive to find the right balance between efficiency and following idioms that are sometimes considered indisputable in other programming paradigms.

Remember that Go is a unique language designed for many characteristics, including simplicity. However, if we find a need for getters and setters or, as mentioned, foresee a future need while guaranteeing forward compatibility, there’s nothing wrong with using them.

Next, we will discuss the problem of overusing interfaces.

# 5: Interface pollution
Interfaces are one of the cornerstones of the Go language when designing and structuring our code. However, like many tools or concepts, abusing them is generally not a good idea. Interface pollution is about overwhelming our code with unnecessary abstractions, making it harder to understand. 

It’s a common mistake made by developers coming from another language with different habits. Before delving into the topic, let’s refresh our minds about Go’s interfaces. Then, we will see when it’s appropriate to use interfaces and when it may be considered pollution.

## Concepts
An interface provides a way to specify the behavior of an object. We use interfaces to create common abstractions that multiple objects can implement. What makes Go interfaces so different is that they are satisfied implicitly. There is no explicit keyword like implements to mark that an object X implements interface Y.

To understand what makes interfaces so powerful, we will dig into two popular ones from the standard library: ```io.Reader``` and ```io.Writer```. The io package provides abstractions for I/O primitives. Among these abstractions, ```io.Reader``` relates to reading data from a data source and ```io.Writer``` to writing data to a target, as represented in figure below.

<img width="1124" height="563" alt="Screenshot 2025-07-30 at 10 46 24 AM" src="https://github.com/user-attachments/assets/1007c4ec-c04b-4ddf-a136-66afce0d90e9" />

The io.Reader contains a single ```Read``` method:
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Custom implementations of the ```io.Reader``` interface should accept a slice of bytes, filling it with its data and returning either the number of bytes read or an error.

On the other hand, ```io.Writer``` defines a single method, ```Write```:
```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Custom implementations of ```io.Writer``` should write the data coming from a slice to a target and return either the number of bytes written or an error. Therefore, both interfaces provide fundamental abstractions:
- ```io.Reader``` reads data from a source.
- ```io.Writer``` writes data to a target.

What is the rationale for having these two interfaces in the language? What is the point of creating these abstractions?
Let’s assume we need to implement a function that should copy the content of one file to another. We could create a specific function that would take as input two ```*os.Files```. Or, we can choose to create a more generic function using ```io.Reader``` and ```io.Writer``` abstractions:

```go
func copySourceToDest(source io.Reader, dest io.Writer) error {
    // ...
}
```

This function would work with ```*os.File``` parameters (as ```*os.File``` implements both ```io.Reader``` and ```io.Writer```) and any other type that would implement these interfaces.
For example, we could create our own ```io.Writer``` that writes to a database, and the code would remain the same. It increases the genericity of the function; hence, its reusability.
Furthermore, writing a unit test for this function is easier because, instead of having to handle files, we can use the strings and bytes packages that provide helpful implementations:

```go
func TestCopySourceToDest(t *testing.T) {
    const input = "foo"
    source := strings.NewReader(input)  // Creates an io.Reader
    dest := bytes.NewBuffer(make([]byte, 0))    // Creates an io.Writer
    // Calls copySourceToDest from a *strings.Reader and a *bytes.Buffer
    err := copySourceToDest(source, dest) 
    if err != nil {
        t.FailNow()
    }

    got := dest.String()
    if got != input {
        t.Errorf("expected: %s, got: %s", input, got)
    }
}
```

In the example, source is a ```*strings.Reader```, whereas dest is a ```*bytes.Buffer```. Here, we test the behavior of copySourceToDest without creating any files.
While designing interfaces, the granularity (how many methods the interface contains) is also something to keep in mind. A known proverb in Go (https://www.youtube.com/watch?v=PAAkCSZUG1c&t=318s) relates to how big an interface should be:

*The bigger the interface, the weaker the abstraction.*
                                            —Rob Pike

Indeed, adding methods to an interface can decrease its level of reusability. ```io.Reader``` and ```io.Writer``` are powerful abstractions because they cannot get any simpler. Furthermore, we can also combine fine-grained interfaces to create higher-level abstractions. This is the case with ```io.ReadWriter```, which combines the ```reader``` and ```writer``` behaviors:

```go
type ReadWriter interface {
    Reader
    Writer
}
```

**NOTE** As Einstein said, “Everything should be made as simple as possible, but no simpler.” Applied to interfaces, this denotes that finding the perfect granularity for an interface isn’t necessarily a straightforward process.

Let’s now discuss common cases where interfaces are recommended.

## When to use interfaces
When should we create interfaces in Go? Let’s look at three concrete use cases where interfaces are usually considered to bring value. Note that the goal isn’t to be exhaustive because the more cases we add, the more they would depend on the context.

However, these three cases should give us a general idea:
- Common behavior
- Decoupling
- Restricting behavior

### COMMON BEHAVIOR
The first option we will discuss is to use interfaces when multiple types implement a common behavior. In such a case, we can factor out the behavior inside an interface.

If we look at the standard library, we can find many examples of such a use case. For example, sorting a collection can be factored out via three methods:
- Retrieving the number of elements in the collection
- Reporting whether one element must be sorted before another
- Swapping two elements

Hence, the following interface was added to the ```sort``` package:
```go
type Interface interface {
    Len() int       // Number of elements
    Less(i, j int) bool // Checks two elements
    Swap(i, j int)  // Swaps two elements
}
```

This interface has a strong potential for reusability because it encompasses the common behavior to sort any collection that is index-based.

Throughout the ```sort``` package, we can find dozens of implementations. If at some point we compute a collection of integers, for example, and we want to sort it, are we necessarily interested in the implementation type? Is it important whether the sorting algorithm is a merge sort or a quicksort? In many cases, we don’t care. Hence, the sorting behavior can be abstracted, and we can depend on the ```sort.Interface```.

Finding the right abstraction to factor out a behavior can also bring many benefits. For example, the ```sort``` package provides utility functions that also rely on sort.Interface, such as checking whether a collection is already sorted. For instance,

```go
func IsSorted(data Interface) bool {
    n := data.Len()
    for i := n - 1; i > 0; i-- {
        if data.Less(i, i-1) {
            return false
        }
    }
    return true
}
```

Because ```sort.Interface``` is the right level of abstraction, it makes it highly valuable. Let’s now see another main use case when using interfaces.

### DECOUPLING
Another important use case is about decoupling our code from an implementation. If we rely on an abstraction instead of a concrete implementation, the implementation itself can be replaced with another without even having to change our code. This is the **Liskov Substitution Principle** (the L in Robert C. Martin’s SOLID design principles).

One benefit of decoupling can be related to unit testing. Let’s assume we want to implement a ```CreateNewCustomer``` method that creates a new customer and stores it. We decide to rely on the concrete implementation directly (let’s say a ```mysql.Store``` struct):

```go
type CustomerService struct {
    store mysql.Store   // Depends on the concrete implementation
}
func (cs CustomerService) CreateNewCustomer(id string) error {
    customer := Customer{id: id}
    return cs.store.StoreCustomer(customer)
}
```

Now, what if we want to test this method? Because ```customerService``` relies on the actual implementation to store a Customer, we are obliged to test it through integration tests, which requires spinning up a MySQL instance (unless we use an alternative technique such as ```go-sqlmock```, but this isn’t the scope of this section). Although integration tests are helpful, that’s not always what we want to do. To give us more flexibility, we should decouple ```CustomerService``` from the actual implementation, which can be done via an interface like so:

```go
type customerStorer interface {
    StoreCustomer(Customer) error   // Creates a storage abstraction
}
type CustomerService struct {
    storer customerStorer   // Decouples CustomerService from the actual implementation
}
func (cs CustomerService) CreateNewCustomer(id string) error {
    customer := Customer{id: id}
    return cs.storer.StoreCustomer(customer)
}
```

Because storing a customer is now done via an interface, this gives us more flexibility in how we want to test the method. For instance, we can
- Use the concrete implementation via integration tests
- Use a mock (or any kind of test double) via unit tests
- Or both

Let’s now discuss another use case: to restrict a behavior.

### RESTRICTING BEHAVIOR
The last use case we will discuss can be pretty counterintuitive at first sight. It’s about restricting a type to a specific behavior. Let’s imagine we implement a custom configuration package to deal with dynamic configuration. We create a specific container for int configurations via an ```IntConfig``` struct that also exposes two methods: ```Get``` and
```Set```. Here’s how that code would look:

```go
type IntConfig struct {
    // ...
}
func (c *IntConfig) Get() int {
    // Retrieve configuration
}
func (c *IntConfig) Set(value int) {
    // Update configuration
}
```

Now, suppose we receive an ```IntConfig``` that holds some specific configuration, such as a threshold. Yet, in our code, we are only interested in retrieving the configuration value, and we want to prevent updating it. How can we enforce that, semantically, this configuration is read-only, if we don’t want to change our configuration package? By creating an abstraction that restricts the behavior to retrieving only a config value:

```go
type intConfigGetter interface {
    Get() int
}
```

Then, in our code, we can rely on intConfigGetter instead of the concrete implementation:
```go
type Foo struct {
    threshold intConfigGetter
}

// Injects the configuration getter
func NewFoo(threshold intConfigGetter) Foo {
    return Foo{threshold: threshold}
}
// Reads the configuration
func (f Foo) Bar() {
    threshold := f.threshold.Get()
    // ...
}
```

In this section, we saw three potential use cases where interfaces are generally considered as bringing value: 
- factoring out a common behavior, 
- creating some decoupling, 
- and restricting a type to a certain behavior. 

Again, this list isn’t exhaustive, but it should give us a general understanding of when interfaces are helpful in Go.

Now, let’s finish this section and discuss the problems with interface pollution.

### Interface pollution
It’s fairly common to see interfaces being overused in Go projects. Perhaps the developer’s background was C# or Java, and they found it natural to create interfaces before concrete types. However, this isn’t how things should work in Go.

As we discussed, interfaces are made to create abstractions. And the main caveat when programming meets abstractions is remembering that abstractions should be discovered, not created. What does this mean? It means we shouldn’t start creating abstractions in our code if there is no immediate reason to do so. We shouldn’t design with interfaces but wait for a concrete need. Said differently, **we should create an interface when we need it, not when we foresee that we could need it**.

What’s the main problem if we overuse interfaces? The answer is that they make the code flow more complex. Adding a useless level of indirection doesn’t bring any value; it creates a worthless abstraction making the code more difficult to read, understand, and reason about. If we don’t have a strong reason for adding an interface and it’s unclear how an interface makes a code better, we should challenge this interface’s purpose. Why not call the implementation directly?

**NOTE** We may also experience performance overhead when calling a method through an interface. It requires a lookup in a hash table’s data structure to find the concrete type an interface points to. But this isn’t an issue in many contexts as the overhead is minimal.

In summary, we should be cautious when creating abstractions in our code— abstractions should be discovered, not created. It’s common for us, software developers, to overengineer our code by trying to guess what the perfect level of abstraction is, based on what we think we might need later. This process should be avoided because, in most cases, it pollutes our code with unnecessary abstractions, making it more complex to read.

*Don’t design with interfaces, discover them.*
—Rob Pike

Let’s not try to solve a problem abstractly but solve what has to be solved now. Last, but
not least, if it’s unclear how an interface makes the code better, we should probably
consider removing it to make our code simpler.

The following section continues with this thread and discusses a common interface mistake: creating interfaces on the producer side.

# 6: Interface on the producer side
We saw in the previous section when interfaces are considered valuable. But Go developers often misunderstand one question: where should an interface live?

Before delving into this topic, let’s make sure the terms we use throughout this section are clear:

- Producer side — An interface defined in the same package as the concrete implementation (see figure 2.4).
- Consumer side — An interface defined in an external package where it’s used (see figure 2.5).

It’s common to see developers creating interfaces on the producer side, alongside the concrete implementation. This design is perhaps a habit from developers having a C# or a Java background. But in Go, in most cases this is not what we should do.

Let’s discuss the following example. Here, we create a specific package to store and retrieve customer data. Meanwhile, still in the same package, we decide that all the calls have to go through the following interface:

```go
package store
type CustomerStorage interface {
    StoreCustomer(customer Customer) error
    GetCustomer(id string) (Customer, error)
    UpdateCustomer(customer Customer) error
    GetAllCustomers() ([]Customer, error)
    GetCustomersWithoutContract() ([]Customer, error)
    GetCustomersWithNegativeBalance() ([]Customer, error)
}
```

We might think we have some excellent reasons to create and expose this interface on the producer side. Perhaps it’s a good way to decouple the client code from the actual implementation. Or, perhaps we can foresee that it will help clients in creating test
doubles. Whatever the reason, this isn’t a best practice in Go.

As mentioned, interfaces are satisfied implicitly in Go, which tends to be a game-changer compared to languages with an explicit implementation. In most cases, the approach to follow is similar to what we described in the previous section: abstractions should be discovered, not created. This means that it’s not up to the producer to force a given abstraction for all the clients. Instead, it’s up to the client to decide whether it needs some form of abstraction and then determine the best abstraction level for its needs.

In the previous example, perhaps one client won’t be interested in decoupling its code. Maybe another client wants to decouple its code but is only interested in the **GetAllCustomers** method. In this case, this client can create an interface with a single method, referencing the Customer struct from the external package:

```go
package client
type customersGetter interface {
    GetAllCustomers() ([]store.Customer, error)
}
```

From a package organization, figure 2.6 shows the result. A couple of things to note:
- Because the **customersGetter** interface is only used in the **client** package, it can remain unexported.
- Visually, in the figure, it looks like circular dependencies. However, there’s no dependency from ```store``` to ```client``` because the interface is satisfied implicitly. This is why such an approach isn’t always possible in languages with an explicit implementation.

The main point is that the ```client``` package can now define the most accurate abstraction for its need (here, only one method). It relates to the concept of the Interface-Segregation Principle (the *I* in SOLID), which states that no client should be forced to depend on methods it doesn’t use. Therefore, in this case, the best approach is to expose the concrete implementation on the producer side and let the client decide how to use it and whether an abstraction is needed.

For the sake of completeness, let’s mention that this approach—interfaces on the producer side—is sometimes used in the standard library. For example, the encoding package defines interfaces implemented by other subpackages such as ```encoding/json``` or ```encoding/binary```. Is the encoding package wrong about this? Definitely not. In this case, the abstractions defined in the ```encoding``` package are used across the standard library, and the language designers knew that creating these abstractions up front was valuable. We are back to the discussion in the previous section: **don’t create an abstraction if you think it might be helpful in an imaginary future** or, at least, if you can’t prove this abstraction is valid.

An interface should live on the consumer side in most cases. However, in particular contexts (for example, when we know - not foresee - that an abstraction will be helpful for consumers), we may want to have it on the producer side. If we do, we should strive to keep it as minimal as possible, increasing its reusability potential and making it more easily composable.

Let’s continue the discussion about interfaces in the context of function signatures.

# 7: Returning Interfaces
While designing a function signature, we may have to return either an interface or a concrete implementation. Let’s understand why **returning an interface is, in many cases, considered a bad practice in Go**.

We just presented why interfaces live, in general, on the consumer side. Figure 2.7 shows what would happen dependency-wise if a function returns an interface instead of a struct. We will see that it leads to issues.

We will consider two packages:
- ```client```, which contains a ```Store``` interface
- ```store```, which contains an implementation of Store

pic

In the ```store``` package, we define an ```InMemoryStore``` struct that implements the ```Store``` interface. Meanwhile, we create a ```NewInMemoryStore``` function to return a ```Store``` interface. There’s a dependency from the implementation package to the ```client``` package in this design, and that may already sound a bit odd.

For example, the ```client``` package can’t call the ```NewInMemoryStore``` function anymore; otherwise, there would be a cyclic dependency. A possible solution could be to call this function from another package and to inject a ```Store``` implementation to ```client```. However, being obliged to do that means that the design should be challenged. Furthermore, what happens if another client uses the ```InMemoryStore``` struct? In that case, perhaps we would like to move the Store interface to another package, or back to the implementation package — but we discussed why, in most cases, this isn’t a best practice. It looks like a code smell.

Hence, in general, returning an interface restricts flexibility because we force all the clients to use one particular type of abstraction. In most cases, we can get inspiration from Postel’s law (https://datatracker.ietf.org/doc/html/rfc761):

*Be conservative in what you do, be liberal in what you accept from others.*
                                                —Transmission Control Protocol

If we apply this idiom to Go, it means
- Returning structs instead of interfaces
- Accepting interfaces if possible

Of course, there are some exceptions. As software engineers, we are familiar with the fact that rules are never true 100% of the time. The most relevant one concerns the error type, an interface returned by many functions. We can also examine another exception in the standard library with the io package:

```go
func LimitReader(r Reader, n int64) Reader {
    return &LimitedReader{r, n}
}
```

Here, the function returns an exported struct, ```io.LimitedReader```. However, the function signature is an interface, ```io.Reader```. What’s the rationale for breaking the rule we’ve discussed so far? The ```io.Reader``` is an up-front abstraction. It’s not one defined by clients, but it’s one that is forced because the language designers knew in advance that this level of abstraction would be helpful (for example, in terms of reusability and composability).

All in all, in most cases, we shouldn’t return interfaces but concrete implementations. Otherwise, it can make our design more complex due to package dependencies and can restrict flexibility because all the clients would have to rely on the same abstraction. Again, the conclusion is similar to the previous sections: if we know (not foresee) that an abstraction will be helpful for clients, we can consider returning an interface. Otherwise, we shouldn’t force abstractions; they should be discovered by clients. If a client needs to abstract an implementation for whatever reason, it can still do that on the client’s side.

In the next section, we will discuss a common mistake related to using any.

# 8: any says nothing
In Go, an interface type that specifies zero methods is known as the empty interface, ```interface{}```. With Go 1.18, the predeclared type any became an alias for an empty interface; hence, all the ```interface{}``` occurrences can be replaced by any. In many cases, any can be considered an overgeneralization; and as mentioned by Rob Pike, it doesn’t convey anything (https://www.youtube.com/watch?v=PAAkCSZUG1c&t=7m36s). Let’s first remind ourselves of the core concepts, and then we can discuss the potential problems.

An any type can hold any value type:

```go
func main() {
    var i any

    i = 42          // An int
    i = "foo"       // A string
    i = struct {    // A struct
        s string
    }{
        s: "bar",
    }
    i = f   // A function
    _ = i   // Assignment to the blank identifier so that the example compiles
}

func f() {}
```

In assigning a value to an ```any``` type, we lose all type information, which requires a type assertion to get anything useful out of the i variable, as in the previous example. Let’s
look at another example, where using any isn’t accurate. In the following, we implement a ```Store``` struct and the skeleton of two methods, ```Get``` and ```Set```. We use these methods to store the different struct types, ```Customer``` and ```Contract```:
