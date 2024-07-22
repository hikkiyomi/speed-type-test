package internal

import "github.com/charmbracelet/lipgloss"

var (
	GRAY   = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	YELLOW = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	RED    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

var styleMapping = map[status]lipgloss.Style{
	STATUS_PENDING: GRAY,
	STATUS_CORRECT: YELLOW,
	STATUS_WRONG:   RED,
}
