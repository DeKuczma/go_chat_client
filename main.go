package main

import (
	"context"
	"log"
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
	log.Println("logging to file")

	// url := url.URL{
	// 	Scheme: "ws",
	// 	Host:   "127.0.0.1:8080",
	// 	Path:   "/ws",
	// }
	// conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer conn.Close()

	log.Println("Connection via websocket established")

	p := tea.NewProgram(NewHub(), tea.WithContext(context.Background()))
	p.Run()

	log.Println("Program run")

	// message := &SettingsMessage{
	// 	Name: "tester QA",
	// }
	// jsonData, err := json.Marshal(message)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// data := string(jsonData)

	// msg := Message{
	// 	MessageType: Settings,
	// 	Data:        data,
	// }
	// log.Printf("sending: %+v", msg)

	// conn.WriteJSON(msg)

	// roomJoin := &RoomOperationMessage{
	// 	Room: "test",
	// }

	// jsonData, err = json.Marshal(roomJoin)

	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	// msg = Message{
	// 	MessageType: Join,
	// 	Data:        string(jsonData),
	// }

	// log.Printf("sending: %+v", msg)
	// conn.WriteJSON(msg)

	// go Send(conn)
	// go Read(conn, p)
}
