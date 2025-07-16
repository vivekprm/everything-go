If we look at below implementation, where we have a slow third party API.

```go 
package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	start := time.Now()
	ctx := context.Background()
	userID := 10
	val, err := fetchUserData(ctx, userID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("result: ", val)
	fmt.Println("time elapsed: ", time.Since(start))
}

func fetchUserData(ctx context.Context, userID int) (int, error) {
	val, err := fetchThirdPartyStuffWhichCanBeSlow()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func fetchThirdPartyStuffWhichCanBeSlow() (int, error) {
	time.Sleep(500 * time.Millisecond)
	return 666, nil
}
```

If we run it, we see it takes around 500ms to ftech our result. So whatever time that API takes we wait for that time. We don't have 
control over it.

We can use context to control this execution and timeout if this API is taking more time. This is how we can modify the existing code.

```go 
package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	start := time.Now()
	ctx := context.Background()
	userID := 10
	val, err := fetchUserData(ctx, userID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("result: ", val)
	fmt.Println("time elapsed: ", time.Since(start))
}

type Response struct {
	value int
	err   error
}

func fetchUserData(ctx context.Context, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	respch := make(chan Response)
	go func() {
		val, err := fetchThirdPartyStuffWhichCanBeSlow()
		respch <- Response{value: val, err: err}
	}()

	for {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("fetching data from thirdparty took too long")
		case resp := <-respch:
			return resp.value, resp.err
		}
	}
}

func fetchThirdPartyStuffWhichCanBeSlow() (int, error) {
	time.Sleep(500 * time.Millisecond)
	return 666, nil
}
```

Context WithValue can be used to pass some data across Goroutines. One usecase could be passing the requestID. In case of errors
we can log this userID and trace the user behaviour.
