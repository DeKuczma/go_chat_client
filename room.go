package main

type Room struct {
	active        bool
	unreadMessage bool
	users         []string
	messages      []ChatMessage
}

type Login struct {
	user string
}

type ChatMessage struct {
	user    string
	message string
}

func (room Room) GetUsers() string {
	var roomUsers string
	for _, user := range room.users {
		roomUsers += user + " "
	}
	return roomUsers
}

func InitRoom() *Room {
	return &Room{
		active:        false,
		unreadMessage: false,
		users:         make([]string, 0),
	}
}
