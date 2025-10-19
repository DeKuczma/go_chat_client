package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func Send(conn *websocket.Conn) {
	time.Sleep(time.Second * 5)

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

func Read(conn *websocket.Conn, p *tea.Program) {

	time.Sleep(time.Second * 5)
	log.Println("ws Reader is running")
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

		log.Printf("Sending msg to tea program %+v\n", msg)
		p.Send(msg)
		log.Println("Sent msg to tea program")
	}
}
