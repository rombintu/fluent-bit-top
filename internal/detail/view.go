package detail

import (
	"fmt"
	"sort"
	"strings"

	"fbtop/internal/shared"

	"charm.land/lipgloss/v2"
)

var titleStyle = lipgloss.NewStyle().Bold(true)

func Render(c shared.ComponentMetrics) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n %s\n", titleStyle.Render(c.ID+" ("+c.Kind.String()+")")))

	keys := make([]string, 0, len(c.Fields))
	for k := range c.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("   %-20s %d\n", k, c.Fields[k]))
	}

	return b.String()
}
