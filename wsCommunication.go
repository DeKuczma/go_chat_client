package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func (h Hub) Send(msg string) tea.Cmd {

	return func() tea.Msg {
		var outcome Message
		switch h.clientState {
		case LogingIn:
			outcome.MessageType = Settings
			structData := &SettingsMessage{Name: msg}
			data, err := json.Marshal(structData)
			if err != nil {
				return tea.Quit()
			}
			outcome.Data = string(data)
		case JoiningRoom:
			outcome.MessageType = Join
			structData := &RoomOperationMessage{Room: msg}
			data, err := json.Marshal(structData)

			log.Printf("%s", data)
			if err != nil {
				return tea.Quit()
			}
			outcome.Data = string(data)
		case WritingMessage:
			outcome.MessageType = TextMessage
			structData := &RoomOperationMessage{Room: msg}
			data, err := json.Marshal(structData)
			if err != nil {
				return tea.Quit()
			}
			outcome.Data = string(data)
		}

		err := h.conn.WriteJSON(outcome)
		if err != nil {
			return tea.Quit()
		}
		return nil
	}
}

func (h *Hub) LeaveRoom() tea.Cmd {

	if h.currentRoomIndex == 0 {
		return nil
	}

	room := h.roomsName[h.currentRoomIndex]
	h.roomsName = append(h.roomsName[:h.currentRoomIndex], h.roomsName[h.currentRoomIndex+1:]...)
	delete(h.rooms, room)
	if h.currentRoomIndex == len(h.roomsName) {
		h.currentRoomIndex--
	}

	return func() tea.Msg {
		outcome := Message{MessageType: Leave}
		structData := &RoomOperationMessage{Room: room}
		data, err := json.Marshal(structData)
		if err != nil {
			return tea.Quit()
		}
		outcome.Data = string(data)

		err = h.conn.WriteJSON(outcome)
		if err != nil {
			return tea.Quit()
		}
		return nil
	}
}

func (h Hub) Disconnect() {
	msg := Message{
		MessageType: Disconnect,
	}

	h.conn.WriteJSON(msg)
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
