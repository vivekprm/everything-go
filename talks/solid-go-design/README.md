# Bad Code
- Rigidity
    - Really hard to change.
- Fragility
    - Does slight change wreaks havoc?
- Immobility
    - Is the code hard to refactor?
- Complexity
    - Is it overly complex for no reason.
- Verbosity
    - Or is it exhausting to use the code?
    - When you look at it do you understand what it tries to do?

# Good Design
Solid Principles described under book called Agile Software Development - By Robert C. Martin
These principles are:

- Single Responsibility Principle
- Open/Closed Principle
- Liskov Substitution Principle
- Interface Segregation Principle
- Dependency Inversion Principle

## Single Responsibility Principle
A class should have one and only one reason to change.

Obviously Golang doesn't have classes however it has much more powerful notion called **Composition**.
The code that has fewer responsibility has fewer reasons to change.

Two words to describe how easy or difficult it is to change a piece of software called **Coupling & Cohesion**.

**Coupling** signifies two things that change together. Movement in one induces movement in the other.

**Cohesion** is the property that descibes pieces of code that naturally attracted to one another.

It starts with Go's package model:

### Package Names
Package name should signify the purpose of the package. Some good package names are defined in standard library e.g.:
- ```net/http```: Gives http client and servers.
- ```os/exec```: Runs external commands.
- ```encoding/json```: Implements encoding and decoding of json document.

When we want to use there packages we use ```import``` declaration which introduces source level coupling between two packages. 

Poorly named packages doesn't signify what it's written for. Some of the Bad package names are:
- package ```server```: Provides server of some kind, but what protocol?
- package ```private```: Provides things that we are not allowed to see?
- package ```common```: Like package util? Dumping ground for miscellenaous. Because they have many responsibilities, they changed frequently and without proper reason.

This principle would be incomplete without mentioning **Go's UNIX Philosophy**: Small sharp tools which combined to solve larger tasks. Often times these tasks are not envisized by the original author. Go packages embody the Go's UNIX Philosophy.


## Open/Closed Principle
Software entities should be open for extension but closed for modification. How can we apply that for Go. Look at below example:

```go
package main

type A struct {
    year int
}

func (a A) Greet() {
    fmt.Println("Hello GolangUK", a.year)
}

type B struct {
    A
}

func (b B) Greet() {
    fmt.Println("Welcome to GolangUK", b.year)
}

func main() {
    var a A
    a.year = 2016
    var b B
    b.year = 2016
    a.Greet()   // Hello GolangUK 2016
    b.Greet()   // Welcome to GolangUK 2016
}
```

B can access embedded type A's private field as though it's defined inside B. So embedding is powerful tool which allows Go's type to be open for extension.

In the second example

```go
package main

type Cat struct {
    Name string
}

func (c Cat) Legs() int {
    return 4
}

func (c Cat) PrintLegs() {
    fmt.Printf("I have %d legs\n", c.Legs())
}

type OctoCat struct {
    Cat
}

func (o OctoCat) Legs() int {
    return 5
}

func main() {
    var octo OctoCat
    fmt.Println(octo.Legs())    // 5
    octo.PrintLegs()            // I have 4 legs
}
```

```PrintLegs``` method returns 4, this is because ```PrintLegs``` is defined on the ```Cat``` type, it takes a ```Cat``` as receiver and so dispatches Cat's Legs method not OctoCat's Legs method.

So ```Cat``` doesn't have any knowlegde of the types it might be embedded into. So it's method set can't be altered by embedding into other types. And so we can say the Go's types are open for extension but closed for modification.

In truth we all know that a method in Go is little more than syntactic sugar around a function with a predeclared formal parameter that being it's receiver. The receiver it exactly what you pass into it.

```go
func PrintLegs(c Cat) {
    fmt.Printf("I have %d legs\n", c.Legs())
}
```

Because Go doesn't support **Function Overloading** OctaCats and not Subtitutable for regular cats and this brings us to the next principle.

## Liskov Substitution Principle
Two types are substitutable if they exhibit behavior such the caller is unable to tell the difference. Now in class-based languages LSP is commonly interpreted as a specification for an abstract base class with various common concrete subtypes. But in Go we don't have classes or Inheritance so substitution can't be implemented like this.

Instead substitution is the purview of Gos interfaces and In Go all types are not required to nominate that they implement a particular interface. Instead any types implements the interface simply by having the matching method set. We say in Go the interfaces are satisfied implicitly rather than explicitly and there is profound impact on how they work.

Well defined interfaces are more likely to be smaller interfaces. The prevailing idiom is that an interface contains only a single method. Small interfaces lead to simple implementations because it's really hard to do anything other than that and this leads to packages composed of simple implementations connected by common behavior.

```go
type Reader interface {
    // Read reads upto len(buf) bytes into buf
    Read(buf []byte) (n int, err error)
}
```

Reader interface is very very simple. Read reads data into the supplied buffer and tells you how many bytes were read and if there was any error encountered during that read.

It seems very simple but it's so powerful because Reader deals with anything that can be expressed as stream of bytes. We can construct Reader over just about anything string, a byte array, stdin, network stream, gzip file, tar gzip file, stdout of a command being executed remotely, all of these. And all these implementations are substitutable for one another because they follow the same simple contract.

So Liskov Substitution Principle applied to Go could be summarized by this "Require no more, promise no less".

## Interface Segregation Principle
Clients should not be forced to depend on methods they don't use. In Go the application of Interface Segregation Principle can refer to the process of isolating the behavior for one Function to do it's Job. 

Now as a concrete example, let's say I have been given the task to write a function that persists some document structure to disk. So I've called it ```Save```

```go
// Save writes the contents of doc to the file f.
func Save(f *os.File, doc *Document) error
```

But this has few problems for example the signature of save precludes the option to write the data to a network location unless ofcourse that network location is mounted as file system somewhere. 

Assuming that network storage is likely to become a requirement later on we would have to change the signature of this function and that would affect all its callers.

```Save``` is also a bit unplesant to test because it operates directly with files on disk. So to verify that this worked under test we have to read out the file so I don't have to read back in the file that I just wrote out. And ofcourse we have to make sure that file is written to temporary location, so that I didn't override something else and I didn't conflict with some other test runs. And ofcourse I have to clean it up at the end.

```os.File``` also defines a lot of methods which are not really relevant to the operation of this ```Save``` method. Like os Files can read directories, can check if a path is symlink and whole bunch of stuff which save is not really interested in.

So it would be really useful if we could write the signature of ```Save``` in a way that told the caller only the bits about ```os.File``` that we are actually interested in.

So one example is that we could use ```io.ReadWriteCloser``` to apply the Interface Segregation Principle to redefine save to take an interface that describes more general file shaped things rather than File.

```go
// Save writes the contents of the doc to the supplied ReadWriteCloser
func Save(rwc io.ReadWriteCloser, doc *Document) error
```

With this change any type that implements ReadWriteCloser can be substituted for os.File. This makes ```Save``` both broader in its application and it clarifies to the callers of save that the only methods from os.File that we are interested in were Read, Write and Close. 

Now as the author of Save, I suddenly no longer have the option to call those extra methods that ```os.File``` provided. I can't cheat anymore because they've been hidden behind the ```ReadWriteCloser``` interface.

We can take InterfaceSegregationPrinciple little bit further, for example it's unlikely that if Save followed the Single Responsibility Principle, it would read the file it just wrote to verify it's content. That should really be the responsibility of a different bit of code to check if the file was written correctly.

So we can narrow the specification at the interface we give to Save, to talk about just writing and closing.

```go
// Save writes the contents of doc to the supplied.
// WriteCloser
func Save(wc io.WriteCloser, doc *Document) error
```

Secondly if we have write Save with a mechanism to close it's stream which we kind of inherited in this desire to make the things still look like a file, raised the question under what circumstances will wc ```WriteCloser``` be closed and maybe you can solve this with documentation but then you've got to read it. But it's possible that ```Save``` might call ```Close``` unconditionally or perhaps close will only be called if save is successful and that kind of ambiguity makes it hard for the caller because as the caller I might want to write additional data out to that stream after I've written that document, may be I want to write several documents that makes it very hard if close will just close that stream straightaway after being used.

So one crude solution would be, we define a new type which embeds ```io.Writer``` and overrides the ```Close``` method preventing ```Save``` from closing the underlying stream.

```go
type NopCloser struct {
    io.Writer
}

// Close has no effect on the underlying writer.
func (c *NopCloser) Close() error { return nil }
```

So here we have ```NopCloser``` with ```Close``` method that just does nothing. But this would potentially be a violation of the Liskov Substitution Principle because now suddenly ```NopCloser``` doesn't close anything anymore. They follow the contract but they don't actually follow the behavior.

A better solution would be to redefine ```Save``` to take only an ```io.Writer```, stripping it completely of the responsibility to do anything but just write data to the stream. 

```go
// Save writes the contents of doc to the supplied
// Writer
func Save(w io.Writer, doc *Document) error
```

By applying this **Interface Segregation Principle** to our ```Save``` function the results simultaneously become a function which is the most specific in terms of it's requirements, it only needs a thing which is Writeable and also most general in it's function. Because now we can use ```Save``` to save data to anything that implements ```io.Writer``` be it a file, a network connection, a byte array whatever.

**A great rule of thumb for Go is accept Interfaces return Structs**


## Dependency Inversion Principle
High Level modules shouldn't depend on low-level modules. Both should depend on Abstractions.
Abstractions shouldn't depend on details. Details should depend on abstractions.

What does dependency inversion mean for in practice for us as Go programmers?
If you've applied all the principles that we've talked about up to this point then your code already going to be factored out into discrete packages, each with a single well-defined responsibility or purpose your code should describe its dependencies in terms of interfaces and those interfaces should be factored to describe only the behavior that they actually require the functions that use them actually require so in short there shouldn't be much left to do at this point if you followed all the design upto this point.

In the context of Go is the structure of your import graph. Now in Go your import graph must be acyclic and a failure to respect this acyclic requirement is ground for a compilation failure but more gravely I think it represents a serious error in design.

All things being equal the import graph of a well-designed Go program should be wide and flat rather than tall and narrow.

<img width="1188" alt="Screenshot 2025-03-24 at 11 09 20 PM" src="https://github.com/user-attachments/assets/ff83041d-ba89-48a3-919a-ba21f609cfad" />

If you have packages whose functions can't operate without enlisting the aid of another that's perhaps the sign that the code is not well factored along package boundaries. So **the dependency inversion principle encourages you to push the responsibility for the specifics as high as possible up in your import graph** leaving low level code to deal in terms of abstractions, interfaces.

- The **Single Responsibility Principle** encourages you to structure your functions, your types and your methods into packages that exhibit natural cohesion. The types belong together the functions serve a single purpose they want to be together.
- The **Open/Closed Principle** encourages you to compose simple types into more complex ones using embedding.
- The **Liskov Substitution Principle** encourages you to express the dependencies between your packages in terms of interfaces not concrete types. By defining small interfaces we can be more confident that the implementations will faithfully satisfy that contract.
- The **Interface Segregation Principle** takes this idea further and encourages you to define functions and methods that depend only on the behavior that they need and if your function requires a parameter of an interface type with a single method then it's more likely that, that function has only one responsibility.
- The **Dependency Inversion Principle** encourages you to move the knowledge of the things that your package depends upon from the compile time to run time and we see this as a reduction in the number of import statements in your source file to runtime.

To summarize this whole thing, then probably the **Interfaces let us apply the Solid Principles to Go programs**. Because interfaces let Go programmers describe what their package provides but not how it does it and it's all just another way of saying decoupling which is indeed the goal because software that's loosely coupled is software that's going to be easier to change.
