package main

type MessageType int

const (
	TextMessage MessageType = iota
	Settings
	Join
	Leave
	Disconnect
	ClientJoined
	ClientLeft
)

type IncomingMessage struct {
	Type    MessageType `json:"type"`
	User    string      `json:"user,omitempty"`
	Room    string      `json:"room,omitempty"`
	Clients []string    `json:"clients,omitempty"`
	Message string      `json:"message,omitempty"`
}

type OutcomingMessage struct {
	Type    MessageType `json:"type"`
	UserId  uint32      `json:"-"`
	User    string      `json:"user,omitempty"`
	Room    string      `json:"room,omitempty"`
	Message string      `json:"message,omitempty"`
}
