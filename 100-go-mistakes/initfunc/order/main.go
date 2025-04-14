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