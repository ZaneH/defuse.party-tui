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

type BigButtonModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	isHolding   bool
	holdSent    bool
	stripColor  pb.Color
	message     string
	messageType string
}

func NewBigButtonModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *BigButtonModule {
	return &BigButtonModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
	}
}

func (m *BigButtonModule) Init() tea.Cmd {
	return nil
}

func (m *BigButtonModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "t", "T":
			return m, m.sendTap()
		case "h", "H":
			if m.isHolding && m.holdSent {
				return m, nil
			}
			return m, m.sendHold()
		case "r", "R":
			if !m.isHolding {
				return m, nil
			}
			return m, m.sendRelease()
		case "esc":
			m.isHolding = false
			m.holdSent = false
			m.stripColor = pb.Color_UNKNOWN
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

func (m *BigButtonModule) sendTap() tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetBigButtonState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no button state")}
		}

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_BigButtonInput{
				BigButtonInput: &pb.BigButtonInput{
					PressType: pb.PressType_TAP,
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if result.GetStrike() {
			m.message = "STRIKE!"
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

func (m *BigButtonModule) sendHold() tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetBigButtonState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no button state")}
		}

		m.isHolding = true
		m.holdSent = true

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_BigButtonInput{
				BigButtonInput: &pb.BigButtonInput{
					PressType: pb.PressType_HOLD,
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			m.isHolding = false
			m.holdSent = false
			return ModuleResultMsg{Err: err}
		}

		if stripResult := result.GetBigButtonInputResult(); stripResult != nil {
			m.stripColor = stripResult.GetStripColor()
		}

		if result.GetStrike() {
			m.isHolding = false
			m.holdSent = false
			m.message = "STRIKE!"
			m.messageType = "error"
			m.stripColor = pb.Color_UNKNOWN
		}

		return ModuleResultMsg{Result: result}
	}
}

func (m *BigButtonModule) sendRelease() tea.Cmd {
	return func() tea.Msg {
		if !m.isHolding {
			return ModuleResultMsg{Err: fmt.Errorf("not holding")}
		}

		state := m.mod.GetBigButtonState()
		if state == nil {
			m.isHolding = false
			m.holdSent = false
			m.stripColor = pb.Color_UNKNOWN
			return ModuleResultMsg{Err: fmt.Errorf("no button state")}
		}

		m.isHolding = false
		m.holdSent = false
		releaseTime := int64(0)

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_BigButtonInput{
				BigButtonInput: &pb.BigButtonInput{
					PressType:        pb.PressType_RELEASE,
					ReleaseTimestamp: releaseTime,
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			m.stripColor = pb.Color_UNKNOWN
			return ModuleResultMsg{Err: err}
		}

		m.stripColor = pb.Color_UNKNOWN

		if result.GetStrike() {
			m.message = "STRIKE!"
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

func (m *BigButtonModule) View() string {
	state := m.mod.GetBigButtonState()
	if state == nil {
		return styles.Error.Render("No button state available (backend issue)")
	}

	buttonColor := state.GetButtonColor()
	label := state.GetLabel()

	colorName := buttonColorToString(buttonColor)
	colorStyle := buttonColorToStyle(buttonColor)

	boxWidth := 21

	labelPadded := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render(label)
	colorPadded := lipgloss.NewStyle().Width(boxWidth).Align(lipgloss.Center).Render("Color: " + colorName)

	buttonDisplay := lipgloss.JoinVertical(
		lipgloss.Center,
		colorStyle.Render("    ┌─────────────┐    "),
		colorStyle.Render("   ╱               ╲   "),
		colorStyle.Render("  ╱                 ╲  "),
		colorStyle.Render(" ╱                   ╲ "),
		colorStyle.Render("│"+labelPadded+"│"),
		colorStyle.Render("│"+colorPadded+"│"),
		colorStyle.Render(" ╲                   ╱ "),
		colorStyle.Render("  ╲                 ╱  "),
		colorStyle.Render("   ╲               ╱   "),
		colorStyle.Render("    └─────────────┘    "),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("BIG BUTTON"),
		"",
		buttonDisplay,
		"",
	)

	if m.isHolding && m.stripColor != pb.Color_UNKNOWN {
		stripColorName := buttonColorToString(m.stripColor)
		stripStyle := buttonColorToStyle(m.stripColor)
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			"",
			stripStyle.Render(fmt.Sprintf("HOLDING - Strip: %s", stripColorName)),
			styles.Subtitle.Render("Press [R] to release when timer shows correct digit"),
		)
	}

	if m.message != "" {
		if m.messageType == "error" {
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				styles.Error.Render(m.message),
			)
		} else if m.messageType == "success" {
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

func (m *BigButtonModule) ID() string {
	return m.mod.GetId()
}

func (m *BigButtonModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_BIG_BUTTON
}

func (m *BigButtonModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *BigButtonModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}
}

func (m *BigButtonModule) Footer() string {
	hint := "[T] Tap | [H] Hold | [R] Release | [ESC] Back to bomb"
	if m.isHolding {
		hint = "[R] Release | [ESC] Cancel"
	}
	return hint
}

func buttonColorToString(c pb.Color) string {
	switch c {
	case pb.Color_RED:
		return "RED"
	case pb.Color_BLUE:
		return "BLUE"
	case pb.Color_WHITE:
		return "WHITE"
	case pb.Color_YELLOW:
		return "YELLOW"
	case pb.Color_ORANGE:
		return "ORANGE"
	case pb.Color_PINK:
		return "PINK"
	case pb.Color_GREEN:
		return "GREEN"
	default:
		return "UNKNOWN"
	}
}

func buttonColorToStyle(c pb.Color) lipgloss.Style {
	switch c {
	case pb.Color_RED:
		return styles.Red
	case pb.Color_BLUE:
		return styles.Blue
	case pb.Color_WHITE:
		return styles.White
	case pb.Color_YELLOW:
		return styles.Yellow
	case pb.Color_ORANGE:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	case pb.Color_PINK:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("219"))
	case pb.Color_GREEN:
		return styles.Green
	default:
		return styles.Normal
	}
}
