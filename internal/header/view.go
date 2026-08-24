package header

import (
	"fmt"

	"fbtop/internal/shared"

	"charm.land/lipgloss/v2"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	urlStyle   = lipgloss.NewStyle().Faint(true)
	okDot      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnDot    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errDot     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	flowStyle  = lipgloss.NewStyle().Faint(true)
	issueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func Render(state shared.State, url string, width int) string {
	var status string
	if !state.Connected {
		status = errDot.Render("✖ Disconnected")
	} else {
		ps := shared.SummarizePipeline(state.Comps)
		if ps.HasErrors {
			status = errDot.Render("✖ Pipeline errors")
		} else if ps.HasWarnings {
			status = warnDot.Render("▲ Warnings")
		} else if ps.Flowing {
			status = okDot.Render("● Healthy")
		} else if ps.InputRec == 0 && ps.OutputRec == 0 {
			status = warnDot.Render("○ Idle")
		} else {
			status = okDot.Render("● Connected")
		}
	}

	title := titleStyle.Render("fluent-bit-top")
	urlStr := urlStyle.Render(url)
	updated := state.UpdatedAt.Format("15:04:05")

	line1 := fmt.Sprintf(" %s    %s    %s    Updated: %s", title, urlStr, status, updated)

	// Pipeline flow summary
	line2 := ""
	if state.Connected {
		ps := shared.SummarizePipeline(state.Comps)
		inStr := shared.FormatCount(ps.InputRec)
		outStr := shared.FormatCount(ps.OutputRec)
		line2 = fmt.Sprintf(" Flow: %s → %d comp → %s    ", inStr, len(state.Comps), outStr)

		if ps.TotalErrors > 0 {
			line2 += issueStyle.Render(fmt.Sprintf("errors:%d ", ps.TotalErrors))
		}
		if ps.TotalRetries > 0 {
			line2 += issueStyle.Render(fmt.Sprintf("retries:%d ", ps.TotalRetries))
		}
		if ps.TotalDropped > 0 {
			line2 += issueStyle.Render(fmt.Sprintf("dropped:%d ", ps.TotalDropped))
		}

		line2 = flowStyle.Render(line2)
	}

	return lipgloss.NewStyle().Width(width).Render(line1 + "\n" + line2)
}
