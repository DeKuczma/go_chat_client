package main

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	url := url.URL{
		Scheme: "ws",
		Host:   "127.0.0.1:8080",
		Path:   "/ws",
	}
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("Connextion via websocket established")
}
