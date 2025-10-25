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

func (room Room) GetUsers(maxHeight int) string {

	users := room.users
	if len(users) > maxHeight {
		users = users[0:maxHeight]
	}
	return lipgloss.JoinVertical(lipgloss.Top, users...)
}

func InitRoom() *Room {
	return &Room{
		active:        false,
		unreadMessage: false,
		users:         make([]string, 0),
	}
}
