package modules

import (
	"context"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking/pkg/proto"
)

const (
	FLASH_DURATION = 0.3
	SEQUENCE_DELAY = 0.75
	SEQUENCE_PAUSE = 2.0
)

type SimonModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	state    *pb.SimonState
	sequence []pb.Color

	showingSequence bool
	isAnimating     bool

	startTime        time.Time
	lastFlashedIndex int

	message     string
	messageType string
}

func NewSimonModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *SimonModule {
	m := &SimonModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
		startTime: time.Now(),
	}

	m.state = mod.GetSimonState()
	if m.state != nil {
		m.sequence = m.state.GetCurrentSequence()
		m.showingSequence = true
	}

	return m
}

func (m *SimonModule) Init() tea.Cmd {
	return tea.Tick(TICK_INTERVAL, func(t time.Time) tea.Msg {
		return SimonTickMsg{Time: t}
	})
}

type SimonTickMsg struct {
	Time time.Time
}

func (m *SimonModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SimonTickMsg:
		m.updateAnimation()
		return m, tea.Tick(TICK_INTERVAL, func(t time.Time) tea.Msg {
			return SimonTickMsg{Time: t}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			return m, m.pressColor(pb.Color_RED)
		case "down", "j":
			return m, m.pressColor(pb.Color_GREEN)
		case "left", "h":
			return m, m.pressColor(pb.Color_YELLOW)
		case "right", "l":
			return m, m.pressColor(pb.Color_BLUE)
		case "r", "R":
			return m, m.pressColor(pb.Color_RED)
		case "g", "G":
			return m, m.pressColor(pb.Color_GREEN)
		case "b", "B":
			return m, m.pressColor(pb.Color_BLUE)
		case "y", "Y":
			return m, m.pressColor(pb.Color_YELLOW)
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

func (m *SimonModule) updateAnimation() {
	if !m.showingSequence || len(m.sequence) == 0 {
		m.isAnimating = false
		return
	}

	elapsed := time.Since(m.startTime).Seconds()

	fullCycleTime := float64(len(m.sequence))*SEQUENCE_DELAY + SEQUENCE_PAUSE
	normalizedTime := math.Mod(elapsed, fullCycleTime)

	if normalizedTime < float64(len(m.sequence))*SEQUENCE_DELAY {
		currentIndex := int(normalizedTime / SEQUENCE_DELAY)
		timeInStep := math.Mod(normalizedTime, SEQUENCE_DELAY)

		if timeInStep < FLASH_DURATION && !m.isAnimating {
			if currentIndex != m.lastFlashedIndex {
				m.isAnimating = true
				m.lastFlashedIndex = currentIndex
			}
		} else if timeInStep >= FLASH_DURATION {
			if currentIndex == m.lastFlashedIndex {
				m.isAnimating = false
				m.lastFlashedIndex = -1
			}
		}
	} else {
		m.isAnimating = false
		m.lastFlashedIndex = -1
	}
}

func (m *SimonModule) pressColor(color pb.Color) tea.Cmd {
	return func() tea.Msg {
		m.showingSequence = false
		m.isAnimating = false
		m.lastFlashedIndex = -1

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_SimonInput{
				SimonInput: &pb.SimonInput{
					Color: color,
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if result.GetSolved() {
			m.message = "Module solved!"
			m.messageType = "success"
			m.showingSequence = false
		} else if result.GetStrike() {
			m.message = "STRIKE!"
			m.messageType = "error"
			if simonResult := result.GetSimonInputResult(); simonResult != nil {
				displaySeq := simonResult.GetDisplaySequence()
				if len(displaySeq) > 0 {
					m.sequence = displaySeq
				}
			}
			m.showingSequence = true
			m.startTime = time.Now()
			m.lastFlashedIndex = -1
			m.isAnimating = false
		} else if simonResult := result.GetSimonInputResult(); simonResult != nil {
			displaySeq := simonResult.GetDisplaySequence()
			if len(displaySeq) > 0 {
				m.sequence = displaySeq
			}

			if simonResult.GetHasFinishedSeq() {
				m.showingSequence = true
				m.startTime = time.Now()
				m.lastFlashedIndex = -1
				m.isAnimating = false
			}
			m.message = ""
			m.messageType = ""
		}

		return ModuleResultMsg{Result: result}
	}
}

func (m *SimonModule) View() string {
	title := styles.Title.Render("SIMON SAYS")

	buttonSize := 12

	redButton := m.renderButton(pb.Color_RED, "RED", buttonSize)
	blueButton := m.renderButton(pb.Color_BLUE, "BLUE", buttonSize)
	greenButton := m.renderButton(pb.Color_GREEN, "GREEN", buttonSize)
	yellowButton := m.renderButton(pb.Color_YELLOW, "YELLOW", buttonSize)

	topRow := lipgloss.JoinHorizontal(lipgloss.Center,
		"",
		redButton,
		"",
	)

	middleRow := lipgloss.JoinHorizontal(lipgloss.Center,
		yellowButton,
		"",
		blueButton,
	)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Center,
		"",
		greenButton,
		"",
	)

	buttons := lipgloss.JoinVertical(
		lipgloss.Center,
		topRow,
		middleRow,
		bottomRow,
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
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
		} else if m.messageType == "success" {
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				"",
				styles.Success.Render(m.message),
			)
		}
	}

	return lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(content)
}

func (m *SimonModule) renderButton(color pb.Color, label string, size int) string {
	bgColor := m.getButtonColor(color)
	borderColor := m.getButtonBorderColor(color)

	litBgColor := m.getLitButtonColor(color)

	if m.showingSequence && len(m.sequence) > 0 {
		elapsed := time.Since(m.startTime).Seconds()
		fullCycleTime := float64(len(m.sequence))*SEQUENCE_DELAY + SEQUENCE_PAUSE
		normalizedTime := math.Mod(elapsed, fullCycleTime)

		if normalizedTime < float64(len(m.sequence))*SEQUENCE_DELAY {
			currentIndex := int(normalizedTime / SEQUENCE_DELAY)
			timeInStep := math.Mod(normalizedTime, SEQUENCE_DELAY)

			if currentIndex < len(m.sequence) && m.sequence[currentIndex] == color && timeInStep < FLASH_DURATION {
				bgColor = litBgColor
				borderColor = litBgColor
			}
		}
	}

	button := lipgloss.NewStyle().
		Width(size).
		Height(size/2).
		Background(bgColor).
		Foreground(lipgloss.Color("#000000")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 0).
		Align(lipgloss.Center).
		Render(label)

	return button
}

func (m *SimonModule) getButtonColor(color pb.Color) lipgloss.Color {
	switch color {
	case pb.Color_RED:
		return lipgloss.Color("#8B0000")
	case pb.Color_BLUE:
		return lipgloss.Color("#00008B")
	case pb.Color_GREEN:
		return lipgloss.Color("#006400")
	case pb.Color_YELLOW:
		return lipgloss.Color("#8B8B00")
	default:
		return lipgloss.Color("#333333")
	}
}

func (m *SimonModule) getButtonBorderColor(color pb.Color) lipgloss.Color {
	switch color {
	case pb.Color_RED:
		return lipgloss.Color("#FF6B6B")
	case pb.Color_BLUE:
		return lipgloss.Color("#4DABF7")
	case pb.Color_GREEN:
		return lipgloss.Color("#6BCB77")
	case pb.Color_YELLOW:
		return lipgloss.Color("#FFD93D")
	default:
		return lipgloss.Color("#555555")
	}
}

func (m *SimonModule) getLitButtonColor(color pb.Color) lipgloss.Color {
	switch color {
	case pb.Color_RED:
		return lipgloss.Color("#FF6B6B")
	case pb.Color_BLUE:
		return lipgloss.Color("#4DABF7")
	case pb.Color_GREEN:
		return lipgloss.Color("#6BCB77")
	case pb.Color_YELLOW:
		return lipgloss.Color("#FFD93D")
	default:
		return lipgloss.Color("#FFFFFF")
	}
}

func (m *SimonModule) ID() string {
	return m.mod.GetId()
}

func (m *SimonModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_SIMON
}

func (m *SimonModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *SimonModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}

	newState := mod.GetSimonState()
	if newState != nil {
		m.state = newState
		m.sequence = newState.GetCurrentSequence()
	}
}

func (m *SimonModule) Footer() string {
	return "[↑/↓/←/→] or [R/G/B/Y] Press button | [ESC] Back to bomb"
}
