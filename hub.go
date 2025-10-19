package main

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

type keyMapping = struct {
	enter,
	quit,
	joinRoom,
	leaveJoinRoom,
	leaveRoom,
	switchNextRoom,
	switchPrevRoom key.Binding
}

type ClientState int

const MaxMessages int = 20

const (
	LogingIn = iota
	JoiningRoom
	WritingMessage
	LeaveRoom
	LeaveHub
)

type Hub struct {
	username         string
	rooms            map[string]*Room
	roomsName        []string
	keyMap           keyMapping
	clientState      ClientState
	input            textinput.Model
	currentRoomIndex int
	conn             *websocket.Conn
}

func NewHub() Hub {
	h := Hub{
		rooms: make(map[string]*Room),

		//Declare key mapping
		keyMap: keyMapping{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("esc", "quit")),
			enter: key.NewBinding(
				key.WithKeys("enter"),
			),
			joinRoom: key.NewBinding(
				key.WithKeys("ctrl+j"),
				key.WithHelp("ctrl+j", "join room"),
			),
			leaveJoinRoom: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "go back to writing messages"),
			),
			switchNextRoom: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "switch to next room"),
			),
			switchPrevRoom: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "switch to prev room"),
			),
			leaveRoom: key.NewBinding(
				key.WithKeys("ctrl+q"),
				key.WithHelp("ctrl+q", "leave current room"),
			),
		},
		clientState:      LogingIn,
		currentRoomIndex: 0,
	}

	//init general room
	h.rooms["general"] = InitRoom()
	h.roomsName = append(h.roomsName, "general")

	//declare input field
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 50
	ti.Prompt = ""

	h.input = ti
	return h
}

func (h Hub) Init() tea.Cmd {

	return nil
}

func (h Hub) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		return h.HandleKeyMsg(m)
	case IncommingMessage:
		log.Printf("Processing incomming msg %+v\n", m)
		room, ok := h.rooms[m.Room]
		if !ok {
			return h, nil
		}

		room.messages = append(room.messages, ChatMessage{
			user:    m.Name,
			message: m.Message,
		})

		if len(room.messages) > MaxMessages {
			room.messages = room.messages[1:]
		}
	}
	return h, nil
}

func (h Hub) View() string {
	m := "Chat!!! \n\n"

	switch h.clientState {
	case LogingIn:
		m += "Provide username: "
	case JoiningRoom:
		m += "Current rooms: "
		m += h.GetJoinedRooms() + "\n\n"
		m += "Provide room name to join: "
	case WritingMessage:
		m += "Current rooms: "
		m += h.GetJoinedRooms() + "\n\n"

		m += "Active room: " + h.roomsName[h.currentRoomIndex] + "\n"
		m += "Room users: " + h.rooms[h.roomsName[h.currentRoomIndex]].GetUsers() + "\n"
		room := h.rooms[h.roomsName[h.currentRoomIndex]]
		if len(room.messages) > 0 {
			m += "Room messages: \n"
			for _, val := range room.messages {
				m += val.user + ": " + val.message + "\n"
			}
		}

		m += "\nWrite message to " + h.roomsName[h.currentRoomIndex] + ": "
	}
	m += h.input.View()
	return m
}

func (h Hub) GetJoinedRooms() string {
	var joinedRooms string
	for _, val := range h.roomsName {
		joinedRooms += val + " "
	}
	return joinedRooms
}

func (h Hub) HandleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, h.keyMap.quit):
		h.clientState = ClientState(Disconnect)
		h.Disconnect()
		return h, tea.Quit
	case key.Matches(msg, h.keyMap.enter):
		return h.HandleInputSubmit()
	case key.Matches(msg, h.keyMap.joinRoom):
		if h.clientState == WritingMessage {
			h.clientState = ClientState(JoiningRoom)
		}
	case key.Matches(msg, h.keyMap.leaveJoinRoom):
		if h.clientState == JoiningRoom {
			h.clientState = ClientState(WritingMessage)
		}
	case key.Matches(msg, h.keyMap.leaveRoom):
		cmd = h.LeaveRoom()
	case key.Matches(msg, h.keyMap.switchNextRoom):
		if h.currentRoomIndex != len(h.roomsName)-1 {
			h.currentRoomIndex++
		}
	case key.Matches(msg, h.keyMap.switchPrevRoom):
		if h.currentRoomIndex != 0 {
			h.currentRoomIndex--
		}
	default:
		h.input, cmd = h.input.Update(msg)
	}

	return h, cmd
}

func (h Hub) HandleInputSubmit() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if len(h.input.Value()) > 0 {
		switch h.clientState {
		case LogingIn:
			h.username = h.input.Value()
			h.input.Reset()
			cmd = h.Send(h.username)
			h.clientState = ClientState(WritingMessage)

			room := h.rooms["general"]
			room.users = append(room.users, h.username)
		case JoiningRoom:
			roomName := h.input.Value()
			h.input.Reset()

			exists := false
			index := 0

			for i, val := range h.roomsName {
				if val == roomName {
					exists = true
					index = i
					break
				}
			}

			if exists {
				h.currentRoomIndex = index
			} else {
				room := InitRoom()
				room.users = append(room.users, h.username)
				h.rooms[roomName] = room
				h.roomsName = append(h.roomsName, roomName)
				h.currentRoomIndex = len(h.roomsName) - 1
				cmd = h.Send(roomName)
			}

			h.clientState = ClientState(WritingMessage)
		case WritingMessage:
			inputMsg := h.input.Value()
			h.input.Reset()
			cmd = h.Send(inputMsg)
		}
	}
	return h, cmd
}
