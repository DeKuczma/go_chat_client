package main

type MessageType int

const (
	TextMessage MessageType = iota
	Settings
	Join
	Leave
	Disconnect
)

type Message struct {
	MessageType MessageType `json:"messageType"`
	Data        string      `json:"data"`
}

type SettingsMessage struct {
	Name string `json:"name"`
}

type SendMessage struct {
	Room    string `json:"room"`
	Message string `json:"message"`
}

type RoomOperationMessage struct {
	Room string `json:"room"`
}

type IncommingMessage struct {
	Room    string `json:"room"`
	Name    string `json:"name"`
	Message string `json:"message"`
}
