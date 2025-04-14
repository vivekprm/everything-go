package redis

import "fmt"

// imports

func init() {
	fmt.Println("redis.go init function")
}
func Store(key, value string) error {
	fmt.Printf("key: %s, value: %s", key, value)
	return nil
}
