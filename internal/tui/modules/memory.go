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

type MemoryModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	message     string
	messageType string
}

func NewMemoryModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *MemoryModule {
	return &MemoryModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
	}
}

func (m *MemoryModule) Init() tea.Cmd {
	return nil
}

func (m *MemoryModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4":
			return m, m.pressButton(int(msg.String()[0] - '1'))
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

func (m *MemoryModule) pressButton(index int) tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetMemoryState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no memory state")}
		}

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_MemoryInput{
				MemoryInput: &pb.MemoryInput{
					ButtonIndex: int32(index),
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

			if memResult := result.GetMemoryInputResult(); memResult != nil {
				if memState := memResult.GetMemoryState(); memState != nil {
					m.mod.State = &pb.Module_MemoryState{MemoryState: memState}
				}
			}
		}

		return ModuleResultMsg{Result: result}
	}
}

func (m *MemoryModule) View() string {
	state := m.mod.GetMemoryState()
	if state == nil {
		return styles.Error.Render("No memory state available")
	}

	screenNumber := state.GetScreenNumber()
	displayedNumbers := state.GetDisplayedNumbers()
	stage := int(state.GetStage())

	screenStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(1, 2).
		Width(6).
		Height(3)

	screenContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Subtitle.Render(fmt.Sprintf("%d", screenNumber)),
	)

	buttons := make([]string, 4)
	for i := 0; i < 4; i++ {
		num := ""
		if i < len(displayedNumbers) {
			num = fmt.Sprintf("%d", displayedNumbers[i])
		}

		buttonStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Width(6).
			Height(3)

		buttonContent := lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Help.Render(fmt.Sprintf("[%d]", i+1)),
			lipgloss.NewStyle().Bold(true).Render(num),
		)

		buttons[i] = buttonStyle.Render(buttonContent)
	}

	buttonsRow := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)

	var stageIndicators []string
	for i := 5; i >= 1; i-- {
		if i < stage {
			stageIndicators = append(stageIndicators, styles.Success.Render("●"))
		} else {
			stageIndicators = append(stageIndicators, styles.Help.Render("○"))
		}
	}

	stageColumn := lipgloss.JoinVertical(
		lipgloss.Center,
		stageIndicators...,
	)

	buttonSection := lipgloss.JoinHorizontal(
		lipgloss.Center,
		buttonsRow,
		lipgloss.NewStyle().Width(2).Render(""),
		stageColumn,
	)

	screenSection := lipgloss.JoinHorizontal(
		lipgloss.Left,
		screenStyle.Render(screenContent),
	)

	mainContent := lipgloss.JoinVertical(
		lipgloss.Center,
		screenSection,
		"",
		buttonSection,
	)

	content := lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styles.Title.Render("MEMORY"),
				"",
				mainContent,
			),
		)

	if m.message != "" {
		if m.messageType == "error" {
			content = lipgloss.NewStyle().
				Width(60).
				Align(lipgloss.Center).
				Render(
					lipgloss.JoinVertical(
						lipgloss.Left,
						content,
						styles.Error.Render(m.message),
					),
				)
		} else if m.messageType == "success" {
			content = lipgloss.NewStyle().
				Width(60).
				Align(lipgloss.Center).
				Render(
					lipgloss.JoinVertical(
						lipgloss.Left,
						content,
						styles.Success.Render(m.message),
					),
				)
		}
	}

	return content
}

func (m *MemoryModule) ID() string {
	return m.mod.GetId()
}

func (m *MemoryModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_MEMORY
}

func (m *MemoryModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *MemoryModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}
}

func (m *MemoryModule) Footer() string {
	return "[1-4] Press button | [ESC] Back to bomb"
}
