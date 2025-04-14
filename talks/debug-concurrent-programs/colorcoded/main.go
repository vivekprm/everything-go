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
