package tableview

import (
	"fmt"
	"strings"

	"fbtop/internal/shared"

	"charm.land/lipgloss/v2"
)

var (
	headerStyle   = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Reverse(true)
)

type Params struct {
	Comps     []shared.ComponentMetrics
	PrevComps []shared.ComponentMetrics
	Width     int
	Cursor    int
	Humanize  bool
}

var (
	healthOKStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	healthWarnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	healthErrStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	healthIdleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func Render(p Params) string {
	prevMap := make(map[string]shared.ComponentMetrics, len(p.PrevComps))
	for _, c := range p.PrevComps {
		prevMap[c.ID] = c
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf(" %-4s %-1s %-24s %-7s %10s %10s %s", "#", "", "Component", "Kind", "Rec/s", "Bytes/s", "Extra")))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", p.Width))
	b.WriteString("\n")

	var lastKind shared.Kind = -1
	for i, c := range p.Comps {
		if c.Kind != lastKind {
			b.WriteString("\n")
			b.WriteString(headerStyle.Render(" " + sectionName(c.Kind)))
			b.WriteString("\n")
			lastKind = c.Kind
		}

		// Rates from delta
		recRate, byteRate := "-", "-"
		if prev, ok := prevMap[c.ID]; ok {
			dt := c.Stamp.Sub(prev.Stamp)
			if dt > 0 {
				recRate = shared.FormatRate(shared.Rate(prev.Records(), c.Records(), dt), p.Humanize, false)
				byteRate = shared.FormatRate(shared.Rate(prev.Bytes(), c.Bytes(), dt), p.Humanize, true)
			}
		}

		// Extra column
		extra := "-"
		switch c.Kind {
		case shared.KindOutput:
			var parts []string
			if e := c.Fields["errors"]; e > 0 {
				parts = append(parts, fmt.Sprintf("err:%d", e))
			}
			if r := c.Fields["retries"]; r > 0 {
				parts = append(parts, fmt.Sprintf("ret:%d", r))
			}
			if len(parts) > 0 {
				extra = strings.Join(parts, " ")
			}
		case shared.KindFilter:
			if d := c.Fields["drop_records"]; d > 0 {
				extra = fmt.Sprintf("drop:%d", d)
			}
		}

		// Health badge
		health := shared.ComponentHealth(c, prevMap[c.ID])
		badge := healthBadge(health)

		row := fmt.Sprintf(" %-4d %-1s %-24s %-7s %10s %10s %s", i+1, badge, c.ID, c.Kind, recRate, byteRate, extra)
		if i == p.Cursor {
			row = selectedStyle.Render(row)
		}
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

func sectionName(k shared.Kind) string {
	switch k {
	case shared.KindInput:
		return "SOURCES"
	case shared.KindFilter:
		return "FILTERS"
	case shared.KindOutput:
		return "OUTPUTS"
	default:
		return "ALL"
	}
}

func healthBadge(h shared.Health) string {
	switch h {
	case shared.HealthOK:
		return healthOKStyle.Render("●")
	case shared.HealthWarning:
		return healthWarnStyle.Render("▲")
	case shared.HealthError:
		return healthErrStyle.Render("✖")
	case shared.HealthIdle:
		return healthIdleStyle.Render("○")
	default:
		return " "
	}
}
