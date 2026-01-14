package modules

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking/pkg/proto"
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

	cutWires map[int32]bool

	message     string
	messageType string
}

func NewWiresModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *WiresModule {
	return &WiresModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
		cutWires:  make(map[int32]bool),
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
			position := int32(msg.String()[0] - '0')
			return m, m.cutWire(position)
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

func (m *WiresModule) cutWire(position int32) tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetWiresState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no wires state")}
		}

		wires := state.GetWires()
		wireExists := false
		for _, wire := range wires {
			if wire.GetPosition() == position {
				wireExists = true
				break
			}
		}

		if !wireExists {
			return ModuleResultMsg{Err: fmt.Errorf("no wire at position %d", position)}
		}

		if m.cutWires[position] {
			return ModuleResultMsg{Err: fmt.Errorf("wire already cut")}
		}

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input:     &pb.PlayerInput_WiresInput{WiresInput: &pb.WiresInput{WirePosition: position}},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		m.cutWires[position] = true

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

	wireMap := make(map[int32]*pb.Wire)
	for _, wire := range wires {
		wireMap[wire.GetPosition()] = wire
	}

	var wireLines []string
	for pos := int32(1); pos <= 6; pos++ {
		wire, exists := wireMap[pos]
		if !exists {
			line := fmt.Sprintf("  %d: ", pos)
			wireLines = append(wireLines, line)
			continue
		}

		colorName := colorToString(wire.GetWireColor())
		colorStyle := colorToStyle(wire.GetWireColor())

		isCut := m.cutWires[pos] || wire.GetIsCut()

		wireDisplay := colorStyle.Render("▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓")
		if isCut {
			wireDisplay = styles.Help.Render("────────────────────")
		}

		line := fmt.Sprintf("  %d: %s  %s", pos, wireDisplay, colorName)
		if isCut {
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
		switch m.messageType {
		case "error":
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				styles.Error.Render(m.message),
			)
		case "success":
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				styles.Success.Render(m.message),
			)
		}
	}

	return lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(content)
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
	if mod.GetSolved() {
		m.mod.Solved = true
	}
}

func (m *WiresModule) Footer() string {
	return "[1-6] Cut wire | [ESC] Back to bomb"
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
