Functions in GO are first class citizens. They are extremely powerful.

# Simple Function

```go 
func add(p1, p2 float64) float64 {
	return p1 + p2
}
```

# Return Multiple Values
```go 
func divide(p1, p2 float64) (float64, error) {
    if p2 == 0 {
        return math.NaN(), errors.New("cannot divide by zero")
    }
    return p1 / p2, nil
}
```

# Variadic Parameter
```go 
func sum(values ...float64) float64 {
    total := 0
    for _, value := range values {
        total += value
    }
    return total
}
```
