package modules

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking/pkg/proto"
)

const (
	VENT_TICK_INTERVAL = time.Second
)

type NeedyVentGasModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	// State from backend
	displayedQuestion  string
	countdownStartedAt int64
	countdownDuration  int32

	message     string
	messageType string
}

func NewNeedyVentGasModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *NeedyVentGasModule {
	m := &NeedyVentGasModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
	}

	state := mod.GetNeedyVentGasState()
	if state != nil {
		m.displayedQuestion = state.GetDisplayedQuestion()
		m.countdownStartedAt = state.GetCountdownStartedAt()
		m.countdownDuration = state.GetCountdownDuration()
	}

	return m
}

type NeedyVentTickMsg struct {
	Time time.Time
}

func (m *NeedyVentGasModule) Init() tea.Cmd {
	return tea.Tick(VENT_TICK_INTERVAL, func(t time.Time) tea.Msg {
		return NeedyVentTickMsg{Time: t}
	})
}

func (m *NeedyVentGasModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NeedyVentTickMsg:
		return m, tea.Tick(VENT_TICK_INTERVAL, func(t time.Time) tea.Msg {
			return NeedyVentTickMsg{Time: t}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			return m, m.sendAnswer(true)
		case "n", "N":
			return m, m.sendAnswer(false)
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

func (m *NeedyVentGasModule) sendAnswer(answer bool) tea.Cmd {
	return func() tea.Msg {
		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_NeedyVentGasInput{
				NeedyVentGasInput: &pb.NeedyVentGasInput{
					Input: answer,
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		// Update state from result
		if ventResult := result.GetNeedyVentGasInputResult(); ventResult != nil {
			if ventState := ventResult.GetNeedyVentGasState(); ventState != nil {
				m.displayedQuestion = ventState.GetDisplayedQuestion()
				m.countdownStartedAt = ventState.GetCountdownStartedAt()
				m.countdownDuration = ventState.GetCountdownDuration()
			}
		}

		if result.GetStrike() {
			m.message = "STRIKE!"
			m.messageType = "error"
		} else {
			m.message = ""
			m.messageType = ""
		}

		return ModuleResultMsg{Result: result}
	}
}

func (m *NeedyVentGasModule) getRemainingTime() int {
	if m.countdownStartedAt == 0 {
		return -1 // Inactive
	}

	now := time.Now().Unix()
	elapsed := now - m.countdownStartedAt
	remaining := int(m.countdownDuration) - int(elapsed)

	if remaining < 0 {
		return 0
	}
	return remaining
}

func (m *NeedyVentGasModule) View() string {
	timer := m.renderTimer()
	question := m.renderQuestion()
	buttons := m.renderButtons()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("NEEDY VENT GAS"),
		"",
		timer,
		"",
		question,
		"",
		buttons,
	)

	if m.message != "" {
		if m.messageType == "error" {
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				"",
				styles.Error.Render(m.message),
			)
		}
	}

	return lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(content)
}

func (m *NeedyVentGasModule) renderTimer() string {
	remaining := m.getRemainingTime()

	redStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	var timerText string
	if remaining < 0 {
		timerText = "--"
	} else {
		timerText = fmt.Sprintf("%02d", remaining)
	}

	timerDisplay := redStyle.Render(timerText)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 2).
		Render(timerDisplay)

	return box
}

func (m *NeedyVentGasModule) renderQuestion() string {
	greenStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5B8C45")).
		Bold(true)

	var questionText string
	if m.displayedQuestion == "" || m.countdownStartedAt == 0 {
		questionText = "WAITING..."
	} else {
		questionText = m.displayedQuestion + "\nY/N"
	}

	questionDisplay := greenStyle.Render(questionText)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5B8C45")).
		Padding(1, 3).
		Align(lipgloss.Center).
		Render(questionDisplay)

	return box
}

func (m *NeedyVentGasModule) renderButtons() string {
	yButton := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Padding(0, 2).
		Render("Y")

	nButton := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Padding(0, 2).
		Render("N")

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		yButton,
		"    ",
		nButton,
	)
}

func (m *NeedyVentGasModule) ID() string {
	return m.mod.GetId()
}

func (m *NeedyVentGasModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_NEEDY_VENT_GAS
}

func (m *NeedyVentGasModule) IsSolved() bool {
	// Needy modules are never "solved" - they just need to be handled when active
	return false
}

func (m *NeedyVentGasModule) UpdateState(mod *pb.Module) {
	state := mod.GetNeedyVentGasState()
	if state != nil {
		m.displayedQuestion = state.GetDisplayedQuestion()
		m.countdownStartedAt = state.GetCountdownStartedAt()
		m.countdownDuration = state.GetCountdownDuration()
	}
}

func (m *NeedyVentGasModule) Footer() string {
	return "[Y] Yes | [N] No | [ESC] Back to bomb"
}
