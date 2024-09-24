package Common

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	LEFT_HALF_CIRCLE  string = string(rune(0xe0b6))
	RIGHT_HALF_CIRCLE string = string(rune(0xe0b4))
	purple                   = lipgloss.Color("#8839ef")
)

var (
	HighlightStyle = lipgloss.NewStyle().
			Background(purple).
			Bold(true).Render

	HighlightStyleFg = lipgloss.NewStyle().
				Foreground(purple).
				Bold(true).Render

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(1).Render
)

func helpView() string {
	helpViewStr := "↑/↓ Move view up/down  • " +
		"c/enter Compile target  • " +
		"s Syntax check  • " +
		"ctrl+c/q Exit program"
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		HighlightStyleFg(LEFT_HALF_CIRCLE),
		HighlightStyle("GO MQL BUILD"),
		HighlightStyleFg(RIGHT_HALF_CIRCLE),
		helpStyle(helpViewStr),
	)
}
