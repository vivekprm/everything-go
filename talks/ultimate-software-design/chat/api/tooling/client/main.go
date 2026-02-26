package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func main() {
	if err := hack1("ws://localhost:3000/connect"); err != nil {
		log.Fatal(err)
	}
}

func hack1(url string) error {
	req := make(http.Header)
	socket, _, err := websocket.DefaultDialer.Dial(url, req)

	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	defer socket.Close()

	// -------------------------------------------

	_, msg, err := socket.ReadMessage()
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	log.Printf("Received message: %s", msg)

	if string(msg) != "HELLO" {
		return fmt.Errorf("unexpected message: %s", msg)
	}

	// -------------------------------------------

	usr := struct{
		Name string `json:"name"`
		ID uuid.UUID `json:"id"`
	}{
		Name: "Vivek",
		ID: uuid.New(),
	}

	data, err := json.Marshal(usr)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := socket.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	// -------------------------------------------

	_, msg, err = socket.ReadMessage()
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	
	log.Printf("Received message: %s", msg)

	return nil
}
