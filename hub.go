package main

import (
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	help             help.Model
	width            int
	height           int
}

func NewHub() Hub {
	h := Hub{
		rooms: make(map[string]*Room),
		help:  help.New(),
		//Declare key mapping
		keyMap: keyMapping{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit")),
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
	ti.CharLimit = 20
	ti.Prompt = "> "

	h.input = ti
	return h
}

func (h Hub) Init() tea.Cmd {

	return nil
}

func (h Hub) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = m.Width
		h.height = m.Height
	case tea.KeyMsg:
		return h.HandleKeyMsg(m)
	case IncomingMessage:
		return h.HandleIncomingMessage(m)
	}
	return h, nil
}

func (h Hub) View() string {
	window := strings.Builder{}
	windowWidth := h.width - windowStyle.GetVerticalBorderSize()
	windowHeight := h.height - windowStyle.GetVerticalBorderSize()
	availableWidth := windowWidth
	availableHeight := windowHeight

	title := titleStyle.Width(availableWidth).Render("\nGo Chat\n")
	window.WriteString(title)
	window.WriteString("\n")
	availableHeight -= 8

	switch h.clientState {
	case LogingIn:
		displayText := "Provide username\n\n" + h.input.View()
		window.WriteString(manageStyle.Width(availableWidth).Height(availableHeight).Render(displayText))
	case JoiningRoom:
		displayText := "Provide room name to join: \n\n" + h.input.View()
		window.WriteString(manageStyle.Width(availableWidth).Height(availableHeight).Render(displayText))
	case WritingMessage:
		usersWidth := 28
		chatWidth := availableWidth - usersWidth - usersStyle.GetHorizontalFrameSize()

		roomUsersPanel := h.GetUsersPanel(usersWidth, availableHeight)
		chatPanel := h.GetChatPanel(chatWidth, availableHeight)
		window.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, chatPanel, roomUsersPanel))
	}
	window.WriteString("\n")

	helpText := topBorderStyle.Width(availableWidth).Render(h.help.ShortHelpView([]key.Binding{
		h.keyMap.joinRoom,
		h.keyMap.leaveRoom,
		h.keyMap.switchNextRoom,
		h.keyMap.switchPrevRoom,
		h.keyMap.quit,
	}))
	window.WriteString(helpText)

	return windowStyle.Width(windowWidth).Height(windowHeight).Render(window.String())
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
		h.currentRoomIndex = min(h.currentRoomIndex+1, len(h.rooms)-1)
		h.rooms[h.roomsName[h.currentRoomIndex]].unreadMessage = false
	case key.Matches(msg, h.keyMap.switchPrevRoom):
		h.currentRoomIndex = max(h.currentRoomIndex-1, 0)
		h.rooms[h.roomsName[h.currentRoomIndex]].unreadMessage = false
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

func (h Hub) HandleIncomingMessage(msg IncomingMessage) (tea.Model, tea.Cmd) {
	log.Printf("Processing incoming msg %+v\n", msg)
	switch msg.Type {
	case TextMessage:
		room, ok := h.rooms[msg.Room]
		if !ok {
			return h, nil
		}

		room.messages = append(room.messages, ChatMessage{
			user:    msg.User,
			message: msg.Message,
		})

		if len(room.messages) > MaxMessages {
			room.messages = room.messages[1:]
		}
		if msg.Room != h.roomsName[h.currentRoomIndex] {
			room.unreadMessage = true
		}
	case ClientLeftRoom:
		room := h.rooms[msg.Room]
		foundAt := 0
		for index, clients := range room.users {
			if clients == msg.User {
				foundAt = index
				break
			}
		}

		room.users = append(room.users[:foundAt], room.users[foundAt+1:]...)
	case ClientJoinedRoom:
		room := h.rooms[msg.Room]
		room.users = append(room.users, msg.User)
	case AllRoomClients:
		room := h.rooms[msg.Room]
		room.users = make([]string, 1, len(msg.Clients))
		room.users[0] = h.username
		for _, client := range msg.Clients {
			if client != h.username {
				room.users = append(room.users, client)
			}
		}
	}
	return h, nil
}

func (h Hub) GetUsersPanel(width, height int) string {
	usersHeader := usersHeaderStyle.Width(width).Render("Users")
	roomUsers := usersStyle.Width(width).Height(height - 3).Align(lipgloss.Center).Render(h.rooms[h.roomsName[h.currentRoomIndex]].GetUsers(height - 3))
	return lipgloss.JoinVertical(lipgloss.Top, usersHeader, roomUsers)
}

func (h Hub) GetChatPanel(width, height int) string {
	height -= 2
	roomHeaderPanel := h.GetRoomsPanel(width)

	inputPanel := messagePanelFooter.Width(width - messagePanelFooter.GetVerticalBorderSize()).Height(1).Render(h.input.View())
	height = height - topBorderStyle.GetHorizontalBorderSize() - activeRoomBorder.GetLeftSize() - activeRoomBorder.GetRightSize()

	room := h.rooms[h.roomsName[h.currentRoomIndex]]
	var messagesPanel string
	var roomMessages []string
	for _, val := range room.messages {
		align := lipgloss.Left
		if strings.Contains(val.user, ("Me")) {
			align = lipgloss.Right
		}
		msg := textStyle.Width(width - textStyle.GetHorizontalBorderSize()).Align(align).Render(val.user + ": " + val.message)
		roomMessages = append(roomMessages, msg)
	}
	if len(roomMessages) >= height {
		roomMessages = roomMessages[len(roomMessages)-height+1:]
	}

	if height-len(roomMessages)-1 > 0 {
		emptyMsg := make([]string, height-len(roomMessages)-1)

		for i := range emptyMsg {
			emptyMsg[i] = textStyle.Width(width - textStyle.GetHorizontalBorderSize()).Align().Render(" ")
		}

		roomMessages = append(emptyMsg, roomMessages...)
	}
	messagesPanel = lipgloss.JoinVertical(lipgloss.Top, roomMessages...)

	return lipgloss.JoinVertical(lipgloss.Top, roomHeaderPanel, messagesPanel, inputPanel)
}

func (h Hub) GetRoomsPanel(width int) string {
	var rooms []string
	var lengths []int
	borderSize := 2

	singleRoomHeader := func(index int, name string) string {
		var border lipgloss.Border
		if index == h.currentRoomIndex {
			border = activeRoomBorder
			if index == 0 {
				border.BottomLeft = "│"
			}
		} else {
			border = inactiveRoomBorder
			if index == 0 {
				border.BottomLeft = "├"
			}
		}
		return roomStyle.Border(border).Render(name)
	}

	createRoomHeader := func(index int) (string, bool) {
		var roomHeader string
		name := h.roomsName[index]
		if width-(len(name)+borderSize) > 5 {
			if name != h.roomsName[h.currentRoomIndex] && h.rooms[name].unreadMessage {
				name += "*"
			}
			lengths = append(lengths, borderSize+len(name))
			roomHeader = singleRoomHeader(index, name)
			width = width - borderSize - len(name)
		} else {
			if width < 5 {
				if (!strings.Contains(rooms[0], "...") && h.currentRoomIndex != 0) || len(rooms) != 1 {
					rooms = append(rooms[0:1], rooms[2:]...)
					width = width + lengths[1]
					lengths = append(lengths[0:1], lengths[2:]...)
				}
			}
			moreText := "..."
			border := inactiveRoomBorder
			width = width - len(moreText) - borderSize

			h.input.SetValue(strconv.Itoa(width))
			if index < h.currentRoomIndex {
				border.BottomLeft = "├"
			} else if width == 0 {
				border.BottomRight = "┤"
			}

			lengths = append(lengths, borderSize+len(moreText))
			roomHeader = roomStyle.Border(border).Render(moreText)
			return roomHeader, false
		}
		return roomHeader, true
	}

	for i := h.currentRoomIndex; i >= 0; i-- {
		roomHeader, cont := createRoomHeader(i)
		rooms = append(rooms, roomHeader)
		if !cont {
			break
		}
	}

	slices.Reverse(rooms)
	slices.Reverse(lengths)

	for i := h.currentRoomIndex + 1; i < len(h.roomsName); i++ {
		roomHeader, cont := createRoomHeader(i)
		if roomHeader != "" {
			rooms = append(rooms, roomHeader)
		}
		if !cont {
			break
		}
	}

	if width > 0 {
		var gapFilling []string

		for i := 0; i < width-1; i++ {
			gapFilling = append(gapFilling, "\n\n─")
		}
		gapFilling = append(gapFilling, "\n\n┐")
		underlineStyle := lipgloss.NewStyle().Foreground(frameColor)
		underline := underlineStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, gapFilling...))
		rooms = append(rooms, underline)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rooms...)

}
