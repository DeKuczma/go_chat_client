package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

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

	log.Println("Connection via websocket established")

	go Send(conn)
	go Read(conn)

	select {}
}

func Send(conn *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		messageType := MessageType(TextMessage)
		line := scanner.Text()
		data := ""
		split := strings.Split(line, " ")
		line = strings.Join(split[1:], " ")
		if split[0] == "disconnect" {
			messageType = MessageType(Disconnect)
		} else if len(split) >= 2 {
			switch split[0] {
			case "settings":
				messageType = MessageType(Settings)
				message := &SettingsMessage{
					Name: line,
				}
				jsonData, err := json.Marshal(message)
				if err != nil {
					fmt.Println(err)
					continue
				}

				data = string(jsonData)
			case "join":
				messageType = MessageType(Join)

				message := &RoomOperationMessage{
					Room: line,
				}
				jsonData, err := json.Marshal(message)

				if err != nil {
					fmt.Println(err)
					continue
				}

				data = string(jsonData)
			case "leave":

				message := &RoomOperationMessage{
					Room: line,
				}

				jsonData, err := json.Marshal(message)

				if err != nil {
					fmt.Println(err)
					continue
				}

				data = string(jsonData)
				messageType = MessageType(Leave)
			default:
				message := &SendMessage{
					Message: line,
					Room:    split[0],
				}
				jsonData, err := json.Marshal(message)
				if err != nil {
					fmt.Println(err)
					continue
				}
				data = string(jsonData)
			}

		}

		msg := Message{
			MessageType: messageType,
			Data:        data,
		}

		conn.WriteJSON(msg)

		if split[0] == "disconnect" {
			conn.Close()
			break
		}
	}
}

func Read(conn *websocket.Conn) {
	for {
		_, recived, err := conn.ReadMessage()

		if err != nil {
			log.Fatal(err)
			return
		}

		var msg IncommingMessage
		err = json.Unmarshal(recived, &msg)

		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("[%s, %s]: %s\n", msg.Room, msg.Name, msg.Message)
	}
}
