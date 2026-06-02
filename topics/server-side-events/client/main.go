package main

import (
	"bufio"
	"log"
	"net/http"
)

func main() {
	resp, err := http.DefaultClient.Get("http://localhost:8088/events")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		log.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading response: %v\n", err)
	}
}