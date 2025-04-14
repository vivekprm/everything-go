package main

import (
	"fmt"

	"gomistakes/initfunc/multiplePackages/redis"
)

func init() {
	fmt.Println("main.go init function")
}

func main() {
	// A dependency on the redis package
	err := redis.Store("foo", "bar")
	if err != nil {
		fmt.Println(err)
	}
}
