package main

import "net/http"

func main() {

    go func() {
       http.ListenAndServe("localhost:6060", nil)
    }()


    go func() {
       for {}
    }()


    for {}
}