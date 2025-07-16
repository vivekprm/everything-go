Let's take a look at invalid regex parsing and try to implement our own Must pattern.

```go 
package main

import (
	"fmt"
	"regexp"
)

func main() {
	// MustComplile handles error for us. Will try to implement our own must pattern.
	// In normal case we should use MustComplile.
	r, err := regexp.Compile("[")
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
}
```

Here we are using Compile deliberately as we want to handle error. But we can use MustComplile usually.
Now we can use our own Must pattern if we don't want to handle error.

We will be using Generics. ```any``` is just ```interface{}```

```go 
package main

import (
	"fmt"
	"regexp"
)

func Must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

func main() {
	r := Must(regexp.Compile("["))
	fmt.Println(r)
}
```

This reduces the boiler plate code adn reduces redundant error checks.

Let's look at another example. Where we want to copy a file. If we see here, there is lot of boiler plate code.

```go
package main

import (
	"io"
	"os"
)

func main() {
	src := "./template.txt"
	dst := "./out/template.txt"

	r, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		panic(err)
	}

	if err := w.Close(); err != nil {
		panic(err)
	}
}
```

We have this large ugly error checking. Let's reduce this error checking leveraging the must function. We can remove the error
and wrap the calls into Must.

```go
package main

import (
	"io"
	"os"
)

func Must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

func main() {
	src := "./template.txt"
	dst := "./out/template.txt"

	r := Must(os.Open(src))
	defer r.Close()

	w := Must(os.Create(dst))
	defer w.Close()

	Must(io.Copy(w, r))

	if err := w.Close(); err != nil {
		panic(err)
	}
}
```

We reduced the error quite a lot. Last error check also can be removed. We can add another function ```checkError```. But if 
we want to refactor this whole thing and put it inside a CopyFile function we can't use the must. As we want to bubble up the
error. So again we have to remove Must. So in case of Copying file it doesn't make sense to use Must.

Place where we can use Must could be during initilization or in case of regex example, where we are 100% sure that regex compiles.
If you have user inputs obviously you don't want to use Must function. Because downside of the Must function is it panics the program.

