package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/events", eventHandler)
	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
 	w.Header().Set("Access-Control-Allow-Origin", "*")
 	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	log.Println("Handler started")

	events := make(chan string)
	go sendEvents(ctx, events)

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			log.Printf("Handler canceled: %v\n", err)
			return

		case data, ok := <-events:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func sendEvents(ctx context.Context, ch chan<- string) {
	defer close(ch)
	i := 0
	for {
		select {
		// If the client closes the connection, the context will be canceled, and we should stop sending events.
		case <-ctx.Done():
			return
		case ch <- fmt.Sprintf("Event %d", i):
		}
		i++
		time.Sleep(2 * time.Second)
	}
}
