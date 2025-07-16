If we look at below server implementation which takes a file name and generates a new filename and e.g. stores it somewhere.

```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Server struct{}

func (s *Server) handleRequest(filename string) error {
	hash := sha256.Sum256([]byte(filename))
	newFileName := hex.EncodeToString(hash[:])

	fmt.Println("New filename is ", newFileName)
	return nil
}

func main() {
	s := &Server{}
	s.handleRequest("cool_pic.jpg")
}
```

But how can we test this functionality, we don't want to create a server to just test this hashing functionality. One thing we can do 
is move file creation in separate function and just test that as below:

```go 
func (s *Server) handleRequest(filename string) error {
	newFileName := hashFilename(filename)
	fmt.Println("New filename is ", newFileName)
	return nil
}

func hashFilename(filename string) string {
	hash := sha256.Sum256([]byte(filename))
	newFileName := hex.EncodeToString(hash[:])
	return newFileName
}
```
 
But that's very mediocre way to do it. What if we want to do a SHA1 or may be prefix with something. How we are going to compose it?
That' where function type composibility kicks in.

You can do it with interface. But you can only if you are not taking any state

```go 
type TrasformFunc func(string) string
```

Now in our server we can take some configuration to descide what logic to apply. We can't change it at runtime.

```go 
type Server struct {
	filenameTransformFunc TrasformFunc
}
```

Now in server what we can do is:
```go 
func (s *Server) handleRequest(filename string) error {
	newFileName := s.filenameTransformFunc(filename)
	fmt.Println("New filename is ", newFileName)
	return nil
}
```

We can pass this filenameTransformFunc as configuration from the client as below:

```go 
func hashFilename(filename string) string {
	hash := sha256.Sum256([]byte(filename))
	newFileName := hex.EncodeToString(hash[:])
	return newFileName
}

func main() {
	s := &Server{
		filenameTransformFunc: hashFilename,
	}
	s.handleRequest("cool_pic.jpg")
}
```

Now if we want to do GG prefix, it's easy to provide that configuration.

We can make it even better, instead of hardcoding the prefix, we can return the tranformFunc from prefixFilename as below:

```go 
func prefixFilename(prefix string) TrasformFunc {
	return func(filename string) string {
		return prefix + filename
	}
}

func main() {
    s := &Server{
		filenameTransformFunc: prefixFilename("BOB_"),
	}
	s.handleRequest("cool_pic.jpg")
}
```

So you can see how we went from something being hardcoded to everything configurable using Function type.
