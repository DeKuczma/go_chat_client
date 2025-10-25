package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func (h Hub) Send(msg string) tea.Cmd {

	return func() tea.Msg {
		var outcome OutcomingMessage
		switch h.clientState {
		case LogingIn:
			outcome.Type = Settings
			outcome.User = msg
		case JoiningRoom:
			outcome.Type = Join
			outcome.Room = msg
		case WritingMessage:
			outcome.Type = TextMessage
			outcome.Room = h.roomsName[h.currentRoomIndex]
			outcome.Message = msg
		}

		val, er := json.Marshal(outcome)

		if er != nil {
			log.Println(er)
			return tea.Quit()
		}

		log.Printf("Sending to server %s \n", string(val))
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
		outcome := OutcomingMessage{Type: Leave,
			Room: room}

		err := h.conn.WriteJSON(outcome)
		if err != nil {
			return tea.Quit()
		}
		return nil
	}
}

func (h Hub) Disconnect() {
	msg := OutcomingMessage{
		Type: Disconnect,
	}

	h.conn.WriteJSON(msg)
}

func Read(ctx context.Context, conn *websocket.Conn, p *tea.Program) {
	log.Println("Started reading message from ws")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, recived, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error sending msg to ws: %s\n", err)
				return
			}

			log.Printf("Recieved: %+v", recived)
			var msg IncomingMessage
			err = json.Unmarshal(recived, &msg)

			if err != nil {
				fmt.Println(err)
				continue
			}

			log.Printf("Sending msg to tea program %+v\n", msg)
			p.Send(msg)
		}
	}
}
