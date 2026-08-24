package footer

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var style = lipgloss.NewStyle().Faint(true).PaddingLeft(1).PaddingTop(1)

func Render(sortCol string, filterKind string, humanize bool, width int) string {
	hum := "on"
	if !humanize {
		hum = "off"
	}

	help := fmt.Sprintf(
		"[Tab] section: %s | [s] sort: %s | [h] humanize: %s | [Enter] detail | [q] quit",
		filterKind,
		sortCol,
		hum,
	)

	return style.Width(width).Render(help)
}
