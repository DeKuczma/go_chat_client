package main

import "github.com/charmbracelet/lipgloss"

var (
	inactiveRoomBorder = roomBorders("┴", "─", "┴")
	activeRoomBorder   = roomBorders("┘", " ", "└")
	textStyle          = lipgloss.NewStyle().Padding(0, 2, 0, 2).Border(messageBorders(), false, true).BorderForeground(frameColor)
	frameColor         = lipgloss.AdaptiveColor{Light: "#4bfd63ff", Dark: "#3a8840ff"}
	roomStyle          = lipgloss.NewStyle().BorderForeground(frameColor)
	defaultStyle       = lipgloss.NewStyle().BorderForeground(frameColor)
	windowStyle        = defaultStyle.Align(lipgloss.Left).Border(lipgloss.RoundedBorder())
	titleStyle         = defaultStyle.Border(lipgloss.NormalBorder(), false, false, true, false).Align(lipgloss.Center)
	manageStyle        = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center)
	topBorderStyle     = defaultStyle.Border(lipgloss.NormalBorder(), true, false, false, false).Align(lipgloss.Center)
	messagePanelFooter = defaultStyle.Border(footerBorders()).Align(lipgloss.Left)
	usersHeaderStyle   = defaultStyle.Border(usersHeaderBorders(), true, true, false, true).AlignHorizontal(lipgloss.Center)
	usersStyle         = messagePanelFooter
)

func roomBorders(left, bottom, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.BottomRight = right
	border.Bottom = bottom
	return border
}

func footerBorders() lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.TopLeft = "├"
	border.TopRight = "┤"
	return border
}

func messageBorders() lipgloss.Border {
	border := lipgloss.NormalBorder()
	border.Left = "│"
	border.Right = "│"
	border.BottomLeft = "|"
	border.BottomRight = "|"
	border.TopLeft = "|"
	border.TopRight = "|"
	return border
}

func usersHeaderBorders() lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = "|"
	border.BottomRight = "|"
	return border
}

func GetLineSpan(text string, width int) int {
	len := len(text)
	if len%width == 0 {
		return len / width
	}
	return len/width + 1
}
