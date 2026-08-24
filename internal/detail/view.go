package detail

import (
	"fmt"
	"sort"
	"strings"

	"fbtop/internal/shared"

	"charm.land/lipgloss/v2"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	okBadge    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	warnBadge  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	errBadge   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	idleBadge  = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true)
	hintStyle  = lipgloss.NewStyle().Faint(true).Italic(true)
	panelStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(0, 1)
)

func Render(c shared.ComponentMetrics, prev shared.ComponentMetrics, width int) string {
	var b strings.Builder

	// Title with health badge
	h := shared.ComponentHealth(c, prev)
	badge := healthBadgeStyled(h)
	b.WriteString(fmt.Sprintf(" %s  %s\n", badge, titleStyle.Render(c.ID+" ("+c.Kind.String()+")")))

	// Health explanation
	b.WriteString(" ")
	b.WriteString(healthExplanation(c, h))
	b.WriteString("\n\n")

	// Fields
	keys := make([]string, 0, len(c.Fields))
	for k := range c.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := c.Fields[k]
		indicator := ""
		if isProblemField(k, v, c.Kind) {
			indicator = errBadge.Render(" ← check this")
		}
		b.WriteString(fmt.Sprintf("   %-22s %d%s\n", k, v, indicator))
	}

	// Diagnostic hints
	hints := componentHints(c, h)
	if len(hints) > 0 {
		b.WriteString("\n")
		for _, h := range hints {
			b.WriteString(fmt.Sprintf("   %s\n", hintStyle.Render("💡 "+h)))
		}
	}

	return panelStyle.Width(width - 2).Render(b.String())
}

func healthBadgeStyled(h shared.Health) string {
	switch h {
	case shared.HealthOK:
		return okBadge.Render("● OK")
	case shared.HealthWarning:
		return warnBadge.Render("▲ WARNING")
	case shared.HealthError:
		return errBadge.Render("✖ ERROR")
	case shared.HealthIdle:
		return idleBadge.Render("○ IDLE")
	default:
		return "?"
	}
}

func healthExplanation(c shared.ComponentMetrics, h shared.Health) string {
	switch h {
	case shared.HealthOK:
		return okBadge.Render("Records flowing normally")
	case shared.HealthError:
		return errBadge.Render("Errors detected — see fields below")
	case shared.HealthWarning:
		return warnBadge.Render("Warnings detected — review below")
	case shared.HealthIdle:
		return idleBadge.Render("No records — check configuration")
	default:
		return ""
	}
}

func isProblemField(key string, val int64, kind shared.Kind) bool {
	if val <= 0 {
		return false
	}
	problemFields := map[string]bool{
		"errors":          true,
		"retries":         true,
		"retries_failed":  true,
		"dropped_records": true,
		"drop_records":    true,
		"drop_bytes":      true,
	}
	return problemFields[key]
}

func componentHints(c shared.ComponentMetrics, h shared.Health) []string {
	var hints []string

	switch c.Kind {
	case shared.KindOutput:
		if c.Fields["errors"] > 0 {
			hints = append(hints, "Check output endpoint: TLS certs, credentials, auth tokens")
			hints = append(hints, "Verify destination host is reachable: check DNS and firewall")
		}
		if c.Fields["retries_failed"] > 0 {
			hints = append(hints, "Increase retry_limit or retry_window in output config")
			hints = append(hints, "Check if destination has rate limits or is rejecting connections")
		}
		if c.Fields["retries"] > 0 && c.Fields["errors"] == 0 {
			hints = append(hints, "Transient failures occurred — monitor or increase retry_limit")
		}
		if c.Fields["dropped_records"] > 0 {
			hints = append(hints, "Increase queue_limit or mem_buf_limit in output config")
			hints = append(hints, "Consider adding a filesystem-backed buffer (fsync: true)")
		}
		if c.Records() == 0 && h == shared.HealthIdle {
			hints = append(hints, "No records reaching this output — check filter match patterns")
			hints = append(hints, "Verify upstream input tags match the output match directive")
		}

	case shared.KindFilter:
		if c.Fields["drop_records"] > 0 {
			total := c.Records() + c.Fields["drop_records"]
			if total > 0 {
				ratio := float64(c.Fields["drop_records"]) / float64(total) * 100
				if ratio > 50 {
					hints = append(hints, fmt.Sprintf("Dropping %.0f%% of records — verify filter conditions", ratio))
				}
			}
		}
		if c.Records() == 0 && h == shared.HealthIdle {
			hints = append(hints, "Check filter match pattern: must match upstream output tags")
			hints = append(hints, "Verify record fields referenced in filter conditions exist")
		}

	case shared.KindInput:
		if c.Records() == 0 && h == shared.HealthIdle {
			hints = append(hints, "Check file path in input config — file may not exist or be empty")
			hints = append(hints, "Verify the input plugin is correctly configured (path, tag, parser)")
			hints = append(hints, "Check Fluent Bit logs for input plugin errors")
		}
	}

	return hints
}
