package tui

import (
	"github.com/ZaneH/defuse.party-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuItem int

const (
	MenuPlayGame MenuItem = iota
	MenuFreePlay
	MenuManual
	MenuQuit
)

var menuItems = []string{
	"PLAY GAME",
	"FREE PLAY",
	"MANUAL",
	"QUIT",
}

func (m *Model) mainMenuView() string {
	var items []string
	for i, item := range menuItems {
		if i == m.menuSelection {
			items = append(items, styles.Active.Render("> "+item))
		} else {
			items = append(items, "  "+item)
		}
	}

	menuContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("DEFUSE.PARTY"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, items...),
	)

	return styles.Center(
		lipgloss.JoinVertical(
			lipgloss.Center,
			menuContent,
		),
		m.width, m.height,
	)
}

func (m *Model) handleMainMenuKeys(key string) (tea.Cmd, bool) {
	handled := true
	switch key {
	case "up", "k":
		if m.menuSelection > 0 {
			m.menuSelection--
		}
	case "down", "j":
		if m.menuSelection < len(menuItems)-1 {
			m.menuSelection++
		}
	case "enter":
		switch MenuItem(m.menuSelection) {
		case MenuPlayGame:
			m.state = StateSectionSelect
			m.sectionSelection = 0
		case MenuFreePlay:
			m.state = StateFreePlayMenu
			m.freePlaySelection = 0
		case MenuManual:
			m.showManualDialog = true
		case MenuQuit:
			return tea.Quit, true
		}
	case "q":
		return tea.Quit, true
	default:
		handled = false
	}
	return nil, handled
}
