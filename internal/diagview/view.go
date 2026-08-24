package diagview

import (
	"fmt"
	"strings"

	"fbtop/internal/shared"

	"charm.land/lipgloss/v2"
)

var (
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	idleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	okStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	panelStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(0, 1)
	titleStyle   = lipgloss.NewStyle().Bold(true)
	hintStyle    = lipgloss.NewStyle().Faint(true).Italic(true)
)

func Render(diags []shared.Diag, width int) string {
	if len(diags) == 0 {
		ok := okStyle.Render("● No issues detected — pipeline is healthy")
		return panelStyle.Width(width - 2).Render(
			titleStyle.Render("Diagnostics") + "\n" + ok,
		)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Diagnostics"))
	b.WriteString("\n")

	for _, d := range diags {
		var badge string
		switch d.Level {
		case shared.HealthError:
			badge = errorStyle.Render("✖ ERROR")
		case shared.HealthWarning:
			badge = warningStyle.Render("▲ WARN ")
		case shared.HealthIdle:
			badge = idleStyle.Render("○ IDLE ")
		default:
			badge = "  OK  "
		}

		comp := ""
		if d.CompID != "" {
			comp = fmt.Sprintf(" [%s]", d.CompID)
		}

		b.WriteString(fmt.Sprintf(" %s %s%s\n", badge, d.Message, comp))
		if d.Hint != "" {
			b.WriteString(fmt.Sprintf("        %s\n", hintStyle.Render("💡 "+d.Hint)))
		}
	}

	return panelStyle.Width(width - 2).Render(b.String())
}
