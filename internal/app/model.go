package app

import (
	"fmt"
	"time"

	"fbtop/internal/api"
	"fbtop/internal/detail"
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

	comps := m.filteredAndSorted()

	parts := []string{
		header.Render(m.state, m.url, m.width),
		tableview.Render(tableview.Params{
			Comps:     comps,
			PrevComps: m.state.PrevComps,
			Width:     m.width,
			Cursor:    m.cursor,
			Humanize:  m.humanize,
		}),
	}

	if m.showDetail && len(comps) > 0 {
		idx := m.cursor
		if idx >= len(comps) {
			idx = len(comps) - 1
		}
		parts = append(parts, detail.Render(comps[idx]))
	}

	parts = append(parts, footer.Render(m.sortCol.String(), m.filterKind.String(), m.humanize, m.width))

	v.SetContent(lipgloss.JoinVertical(lipgloss.Top, parts...))
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
