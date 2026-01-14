package modules

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking-tui/proto"
)

const (
	KNOB_TICK_INTERVAL = time.Second
)

type NeedyKnobModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	// State from backend
	displayedPatternFirstRow  []bool
	displayedPatternSecondRow []bool
	dialDirection             pb.CardinalDirection
	countdownStartedAt        int64
	countdownDuration         int32

	message     string
	messageType string
}

func NewNeedyKnobModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *NeedyKnobModule {
	m := &NeedyKnobModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
	}

	state := mod.GetNeedyKnobState()
	if state != nil {
		m.displayedPatternFirstRow = state.GetDisplayedPatternFirstRow()
		m.displayedPatternSecondRow = state.GetDisplayedPatternSecondRow()
		m.dialDirection = state.GetDialDirection()
		m.countdownStartedAt = state.GetCountdownStartedAt()
		m.countdownDuration = state.GetCountdownDuration()
	}

	return m
}

type NeedyKnobTickMsg struct {
	Time time.Time
}

func (m *NeedyKnobModule) Init() tea.Cmd {
	return tea.Tick(KNOB_TICK_INTERVAL, func(t time.Time) tea.Msg {
		return NeedyKnobTickMsg{Time: t}
	})
}

func (m *NeedyKnobModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NeedyKnobTickMsg:
		return m, tea.Tick(KNOB_TICK_INTERVAL, func(t time.Time) tea.Msg {
			return NeedyKnobTickMsg{Time: t}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return m, m.sendRotate()
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

func (m *NeedyKnobModule) sendRotate() tea.Cmd {
	return func() tea.Msg {
		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_NeedyKnobInput{
				NeedyKnobInput: &pb.NeedyKnobInput{},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		// Update state from result
		if knobResult := result.GetNeedyKnobInputResult(); knobResult != nil {
			if knobState := knobResult.GetNeedyKnobState(); knobState != nil {
				m.displayedPatternFirstRow = knobState.GetDisplayedPatternFirstRow()
				m.displayedPatternSecondRow = knobState.GetDisplayedPatternSecondRow()
				m.dialDirection = knobState.GetDialDirection()
				m.countdownStartedAt = knobState.GetCountdownStartedAt()
				m.countdownDuration = knobState.GetCountdownDuration()
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

func (m *NeedyKnobModule) getRemainingTime() int {
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

func (m *NeedyKnobModule) View() string {
	timer := m.renderTimer()
	dial := m.renderDial()
	leds := m.renderLEDs()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("NEEDY KNOB"),
		"",
		timer,
		"",
		dial,
		"",
		leds,
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

	return content
}

func (m *NeedyKnobModule) renderTimer() string {
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

func (m *NeedyKnobModule) renderDial() string {
	// Dial with arrow only in the direction it's pointing
	var line1, line2, line3 string

	switch m.dialDirection {
	case pb.CardinalDirection_NORTH:
		line1 = "   ▲   "
		line2 = "   ●   "
		line3 = "       "
	case pb.CardinalDirection_EAST:
		line1 = "       "
		line2 = "   ● ► "
		line3 = "       "
	case pb.CardinalDirection_SOUTH:
		line1 = "       "
		line2 = "   ●   "
		line3 = "   ▼   "
	case pb.CardinalDirection_WEST:
		line1 = "       "
		line2 = " ◄ ●   "
		line3 = "       "
	}

	dialContent := fmt.Sprintf("%s\n%s\n%s", line1, line2, line3)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Render(dialContent)

	return box
}

func (m *NeedyKnobModule) renderLEDs() string {
	litStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6BCB77"))

	unlitStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	// Render first row (3 + space + 3)
	row1 := ""
	for i := 0; i < 6; i++ {
		if i == 3 {
			row1 += "   "
		}
		var lit bool
		if i < len(m.displayedPatternFirstRow) {
			lit = m.displayedPatternFirstRow[i]
		}
		if lit {
			row1 += litStyle.Render("●") + " "
		} else {
			row1 += unlitStyle.Render("○") + " "
		}
	}

	// Render second row (3 + space + 3)
	row2 := ""
	for i := 0; i < 6; i++ {
		if i == 3 {
			row2 += "   "
		}
		var lit bool
		if i < len(m.displayedPatternSecondRow) {
			lit = m.displayedPatternSecondRow[i]
		}
		if lit {
			row2 += litStyle.Render("●") + " "
		} else {
			row2 += unlitStyle.Render("○") + " "
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		row1,
		row2,
	)
}

func (m *NeedyKnobModule) ID() string {
	return m.mod.GetId()
}

func (m *NeedyKnobModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_NEEDY_KNOB
}

func (m *NeedyKnobModule) IsSolved() bool {
	// Needy modules are never "solved" - they just need to be handled when active
	return false
}

func (m *NeedyKnobModule) UpdateState(mod *pb.Module) {
	state := mod.GetNeedyKnobState()
	if state != nil {
		m.displayedPatternFirstRow = state.GetDisplayedPatternFirstRow()
		m.displayedPatternSecondRow = state.GetDisplayedPatternSecondRow()
		m.dialDirection = state.GetDialDirection()
		m.countdownStartedAt = state.GetCountdownStartedAt()
		m.countdownDuration = state.GetCountdownDuration()
	}
}

func (m *NeedyKnobModule) Footer() string {
	return "[ENTER] Rotate dial | [ESC] Back to bomb"
}
