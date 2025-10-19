package main

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type keyMapping = struct {
	enter,
	quit,
	joinRoom,
	leaveJoinRoom,
	switchNextRoom,
	switchPrevRoom key.Binding
}

type ClientState int

const MaxMessages int = 20

const (
	LogingIn = iota
	JoiningRoom
	WritingMessage
)

type Hub struct {
	username         string
	logedIn          bool
	rooms            map[string]*Room
	roomsName        []string
	keyMap           keyMapping
	clientState      ClientState
	input            textinput.Model
	currentRoomIndex int
}

func NewHub() Hub {
	h := Hub{
		logedIn: false,
		rooms:   make(map[string]*Room),

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
	var cmd tea.Cmd
	log.Println(msg)
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(m, h.keyMap.quit):
			return h, tea.Quit
		case key.Matches(m, h.keyMap.enter):
			if len(h.input.Value()) > 0 {

				switch h.clientState {
				case LogingIn:
					h.username = h.input.Value()
					h.input.Reset()
					h.clientState = WritingMessage

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
					}

					h.clientState = WritingMessage
				case WritingMessage:
					room := h.rooms[h.roomsName[h.currentRoomIndex]]
					room.messages = append(room.messages, ChatMessage{
						user:    h.username,
						message: h.input.Value(),
					})
					h.input.Reset()
				}

			}
		case key.Matches(m, h.keyMap.joinRoom):
			if h.clientState == WritingMessage {
				h.clientState = JoiningRoom
			}
		case key.Matches(m, h.keyMap.leaveJoinRoom):
			if h.clientState == JoiningRoom {
				h.clientState = WritingMessage
			}
		case key.Matches(m, h.keyMap.switchNextRoom):
			if h.currentRoomIndex != len(h.roomsName)-1 {
				h.currentRoomIndex++
			}
		case key.Matches(m, h.keyMap.switchPrevRoom):
			if h.currentRoomIndex != 0 {
				h.currentRoomIndex--
			}
		default:
			h.input, cmd = h.input.Update(msg)
			return h, cmd
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
		for _, room := range h.rooms {
			if len(room.messages) > 0 {
				m += "Room messages: \n"
				for _, val := range room.messages {
					m += val.user + ": " + val.message + "\n"
				}
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
