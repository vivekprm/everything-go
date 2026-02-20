package main

import (
	"net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

    go func() {
       http.Handle("/metrics", promhttp.Handler())
       http.ListenAndServe(":2112", nil)
    }()

    go func(){
       for{}
    }()

    for {}
}