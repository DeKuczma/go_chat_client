package main

import (
	"context"
	"log"
	"net/url"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

var conn *websocket.Conn

func main() {

	f, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	log.SetOutput(f)

	url := url.URL{
		Scheme: "ws",
		Host:   "127.0.0.1:8080",
		Path:   "/ws",
	}
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	h := NewHub()
	h.conn = conn
	defer h.conn.Close()

	log.Println("Connection via websocket established")

	p := tea.NewProgram(h, tea.WithAltScreen())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go Read(ctx, h.conn, p)

	p.Run()

	log.Println("Program finished running")
}
