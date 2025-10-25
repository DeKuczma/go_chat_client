package main

import "github.com/charmbracelet/lipgloss"

type Room struct {
	active        bool
	unreadMessage bool
	users         []string
	messages      []ChatMessage
}

type ChatMessage struct {
	user    string
	message string
}

func (room Room) GetUsers() string {
	return lipgloss.JoinVertical(lipgloss.Top, room.users...)
}

func InitRoom() *Room {
	return &Room{
		active:        false,
		unreadMessage: false,
		users:         make([]string, 0),
	}
}
