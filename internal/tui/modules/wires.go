package modules

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking-tui/proto"
)

type ModuleResultMsg struct {
	Result *pb.PlayerInputResult
	Err    error
}

type BackToBombMsg struct{}

type WiresModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	message     string
	messageType string
}

func NewWiresModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *WiresModule {
	return &WiresModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
	}
}

func (m *WiresModule) Init() tea.Cmd {
	return nil
}

func (m *WiresModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5", "6":
			wireNum := int(msg.String()[0] - '1')
			return m, m.cutWire(wireNum)
		case "esc":
			return m, func() tea.Msg {
				return BackToBombMsg{}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *WiresModule) cutWire(wireNum int) tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetWiresState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no wires state")}
		}

		wires := state.GetWires()
		if wireNum < 0 || wireNum >= len(wires) {
			return ModuleResultMsg{Err: fmt.Errorf("invalid wire number")}
		}

		wire := wires[wireNum]
		if wire.GetIsCut() {
			return ModuleResultMsg{Err: fmt.Errorf("wire already cut")}
		}

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input:     &pb.PlayerInput_WiresInput{WiresInput: &pb.WiresInput{WirePosition: int32(wireNum)}},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if result.GetStrike() {
			m.message = "STRIKE! Wrong wire!"
			m.messageType = "error"
		} else if result.GetSolved() {
			m.message = "Module solved!"
			m.messageType = "success"
		} else {
			m.message = ""
			m.messageType = ""
		}

		return ModuleResultMsg{Result: result}
	}
}

func (m *WiresModule) View() string {
	state := m.mod.GetWiresState()
	if state == nil {
		return styles.Error.Render("No wires state available")
	}

	wires := state.GetWires()
	if len(wires) == 0 {
		return styles.Subtitle.Render("No wires on this module")
	}

	var wireLines []string
	for i, wire := range wires {
		colorName := colorToString(wire.GetWireColor())
		colorStyle := colorToStyle(wire.GetWireColor())

		wireDisplay := colorStyle.Render("▓▓▓▓▓▓▓▓▓▓▓▓▓")
		if wire.GetIsCut() {
			wireDisplay = styles.Help.Render("────────────────────")
		}

		line := fmt.Sprintf("  %d: %s  %s", i+1, wireDisplay, colorName)
		if wire.GetIsCut() {
			line += " (CUT)"
		}
		wireLines = append(wireLines, line)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("WIRES"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, wireLines...),
		"",
	)

	if m.message != "" {
		if m.messageType == "error" {
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				content,
				styles.Error.Render(m.message),
			)
		} else if m.messageType == "success" {
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				content,
				styles.Success.Render(m.message),
			)
		}
	}

	return content
}

func (m *WiresModule) ID() string {
	return m.mod.GetId()
}

func (m *WiresModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_WIRES
}

func (m *WiresModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *WiresModule) UpdateState(mod *pb.Module) {
	m.mod = mod
}

func (m *WiresModule) Footer() string {
	return "[1-6] Cut wire | [ESC] Back to bomb | [Q]uit"
}

func colorToString(c pb.Color) string {
	switch c {
	case pb.Color_RED:
		return "RED"
	case pb.Color_BLUE:
		return "BLUE"
	case pb.Color_WHITE:
		return "WHITE"
	case pb.Color_BLACK:
		return "BLACK"
	case pb.Color_YELLOW:
		return "YELLOW"
	case pb.Color_GREEN:
		return "GREEN"
	case pb.Color_ORANGE:
		return "ORANGE"
	case pb.Color_PINK:
		return "PINK"
	default:
		return "UNKNOWN"
	}
}

func colorToStyle(c pb.Color) lipgloss.Style {
	switch c {
	case pb.Color_RED:
		return styles.Red
	case pb.Color_BLUE:
		return styles.Blue
	case pb.Color_WHITE:
		return styles.White
	case pb.Color_BLACK:
		return styles.Black
	case pb.Color_YELLOW:
		return styles.Yellow
	case pb.Color_GREEN:
		return styles.Green
	default:
		return styles.Normal
	}
}
