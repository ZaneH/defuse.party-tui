package modules

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking-tui/proto"
)

type UnimplementedModule struct {
	mod        *pb.Module
	width      int
	height     int
	moduleName string
}

func NewUnimplementedModule(mod *pb.Module) *UnimplementedModule {
	return &UnimplementedModule{
		mod:        mod,
		moduleName: moduleTypeName(mod.GetType()),
	}
}

func (m *UnimplementedModule) Init() tea.Cmd {
	return nil
}

func (m *UnimplementedModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, func() tea.Msg {
			return BackToBombMsg{}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *UnimplementedModule) View() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render(m.moduleName),
		"",
		styles.Subtitle.Render("This module type is not yet implemented."),
		"",
		styles.Help.Render("Press [ESC] to return to the bomb."),
	)
	return content
}

func (m *UnimplementedModule) ID() string {
	return m.mod.GetId()
}

func (m *UnimplementedModule) ModuleType() pb.Module_ModuleType {
	return m.mod.GetType()
}

func (m *UnimplementedModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *UnimplementedModule) UpdateState(mod *pb.Module) {
	m.mod = mod
}

func (m *UnimplementedModule) Footer() string {
	return "[ESC] Back to bomb | [Q]uit"
}

func moduleTypeName(t pb.Module_ModuleType) string {
	switch t {
	case pb.Module_WIRES:
		return "WIRES"
	case pb.Module_PASSWORD:
		return "PASSWORD"
	case pb.Module_BIG_BUTTON:
		return "BIG BUTTON"
	case pb.Module_SIMON:
		return "SIMON"
	case pb.Module_KEYPAD:
		return "KEYPAD"
	case pb.Module_WHOS_ON_FIRST:
		return "WHO'S ON FIRST"
	case pb.Module_MEMORY:
		return "MEMORY"
	case pb.Module_MORSE:
		return "MORSE CODE"
	case pb.Module_NEEDY_VENT_GAS:
		return "VENT GAS"
	case pb.Module_NEEDY_KNOB:
		return "KNOB"
	case pb.Module_MAZE:
		return "MAZE"
	default:
		return fmt.Sprintf("UNKNOWN (type %d)", t)
	}
}
