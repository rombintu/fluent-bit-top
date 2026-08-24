package header

import (
	"fmt"

	"fbtop/internal/shared"

	"charm.land/lipgloss/v2"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	urlStyle   = lipgloss.NewStyle().Faint(true)
	dotStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

func Render(state shared.State, url string, width int) string {
	dot := "● Disconnected"
	if state.Connected {
		dot = dotStyle.Render("● Connected")
	}

	title := titleStyle.Render("fbtop")
	urlStr := urlStyle.Render(url)
	updated := state.UpdatedAt.Format("15:04:05")

	header := fmt.Sprintf(" %s    %s    %s\n Components: %d   Updated: %s",
		title, urlStr, dot, len(state.Comps), updated)

	return lipgloss.NewStyle().Width(width).Render(header)
}
