package main

import (
	"errors"
	"fmt"
	"math"
)

func main() {
	fmt.Printf("%f\n", add(2, 6))
	ans, err := divide(6, 2)
	if err != nil {
		fmt.Printf("an error occurred %s\n", err.Error())
	} else {
		fmt.Printf("%f\n", ans)
	}
}

func add(p1, p2 float64) float64 {
	return p1 + p2
}

func divide(p1, p2 float64) (float64, error) {
	if p2 == 0 {
		return math.NaN(), errors.New("cannot divide by zero")
	}
	return p1 / p2, nil
}
