package app

import (
	"fmt"
	"time"

	"fbtop/internal/api"
	"fbtop/internal/detail"
	"fbtop/internal/diagview"
	"fbtop/internal/footer"
	"fbtop/internal/header"
	"fbtop/internal/shared"
	"fbtop/internal/tableview"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type tickMsg struct{}
type errMsg struct{ err error }

type Model struct {
	state      shared.State
	url        string
	interval   time.Duration
	width      int
	height     int
	cursor     int
	humanize   bool
	showDetail bool
	showDiag   bool
	sortCol    shared.SortCol
	sortAsc    bool
	filterKind shared.Kind
	err        error
}

func New(url string, interval time.Duration, humanize bool, sortCol shared.SortCol, sortAsc bool, filterKind shared.Kind) Model {
	return Model{
		url:        url,
		interval:   interval,
		humanize:   humanize,
		showDetail: true,
		showDiag:   true,
		sortCol:    sortCol,
		sortAsc:    sortAsc,
		filterKind: filterKind,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetch(), tick(m.interval))
}

func tick(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return tickMsg{}
	}
}

func (m Model) fetch() tea.Cmd {
	prev := m.state.Comps
	return func() tea.Msg {
		s, err := api.Fetch(m.url)
		if err != nil {
			return errMsg{err}
		}
		s.PrevComps = prev
		return s
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case shared.State:
		m.state = msg
		m.err = nil
		return m, tick(m.interval)
	case errMsg:
		m.err = msg.err
		m.state.Connected = false
		return m, tick(m.interval)
	case tickMsg:
		return m, m.fetch()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filteredAndSorted())-1 {
				m.cursor++
			}
		case "tab": // cycle filter: all→input→filter→output
			m.filterKind = (m.filterKind + 1) % 4
			m.cursor = 0
		case "s":
			m.sortCol = m.sortCol.Next()
		case "S":
			m.sortAsc = !m.sortAsc
		case "h":
			m.humanize = !m.humanize
		case "enter":
			m.showDetail = !m.showDetail
		case "esc":
			m.showDetail = false
		case "d":
			m.showDiag = !m.showDiag
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true

	if m.width == 0 {
		v.SetContent("Initializing...")
		return v
	}
	if m.err != nil {
		v.SetContent(fmt.Sprintf("Error: %v\n\nPress [q] to quit.", m.err))
		return v
	}

	showDetail := m.showDetail && len(m.filteredAndSorted()) > 0

	tableW := m.width
	rightW := 0
	if showDetail && m.width > 100 {
		rightW = m.width * 2 / 5
		tableW = m.width - rightW
	} else if showDetail {
		rightW = m.width / 2
		tableW = m.width - rightW
	}

	comps := m.filteredAndSorted()

	// Header
	headerLine := header.Render(m.state, m.url, m.width)

	// Left pane: table + diagnostics stacked vertically
	leftParts := []string{
		tableview.Render(tableview.Params{
			Comps:     comps,
			PrevComps: m.state.PrevComps,
			Width:     tableW,
			Cursor:    m.cursor,
			Humanize:  m.humanize,
		}),
	}
	if m.showDiag {
		diags := shared.AnalyzeDiagnostics(m.state, m.state.PrevComps)
		leftParts = append(leftParts, diagview.Render(diags, tableW))
	}
	leftContent := lipgloss.JoinVertical(lipgloss.Top, leftParts...)

	// Right pane: detail only
	rightContent := ""
	if showDetail {
		idx := m.cursor
		if idx >= len(comps) {
			idx = len(comps) - 1
		}
		prev := prevComp(comps[idx].ID, m.state.PrevComps)
		rightContent = detail.Render(comps[idx], prev, rightW)
	}

	// Main area
	var main string
	if rightContent != "" {
		main = lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
	} else {
		main = leftContent
	}

	// Footer
	foot := footer.Render(m.sortCol.String(), m.filterKind.String(), m.humanize, m.showDiag, m.width)

	v.SetContent(lipgloss.JoinVertical(lipgloss.Top, headerLine, main, foot))
	return v
}

func (m Model) filteredAndSorted() []shared.ComponentMetrics {
	comps := make([]shared.ComponentMetrics, 0, len(m.state.Comps))
	for _, c := range m.state.Comps {
		if m.filterKind == shared.KindAll || c.Kind == m.filterKind {
			comps = append(comps, c)
		}
	}
	shared.SortBy(comps, m.sortCol, m.sortAsc)
	return comps
}

func prevComp(id string, prevComps []shared.ComponentMetrics) shared.ComponentMetrics {
	for _, c := range prevComps {
		if c.ID == id {
			return c
		}
	}
	return shared.ComponentMetrics{}
}
