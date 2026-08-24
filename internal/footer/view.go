package footer

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var style = lipgloss.NewStyle().Faint(true).PaddingLeft(1).PaddingTop(1)

func Render(sortCol string, filterKind string, humanize bool, showDiag bool, width int) string {
	hum := "on"
	if !humanize {
		hum = "off"
	}
	diag := "off"
	if showDiag {
		diag = "on"
	}

	help := fmt.Sprintf(
		"[Tab] section: %s | [s] sort: %s | [h] humanize: %s | [d] diagnostics: %s | [Enter] detail | [q] quit",
		filterKind,
		sortCol,
		hum,
		diag,
	)

	return style.Width(width).Render(help)
}
