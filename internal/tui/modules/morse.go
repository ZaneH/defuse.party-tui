package modules

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking-tui/proto"
)

const (
	DOT_DURATION  = 0.3
	DASH_DURATION = 0.9
	SYMBOL_PAUSE  = 0.3
	LETTER_PAUSE  = 0.9
	TICK_INTERVAL = 50 * time.Millisecond
)

var morseFrequencies = []float32{
	3.505, 3.515, 3.522, 3.532, 3.535,
	3.542, 3.545, 3.552, 3.555, 3.565,
	3.572, 3.575, 3.582, 3.592, 3.595, 3.600,
}

func frequencyToIndex(freq float32) int32 {
	closestIdx := int32(0)
	minDiff := float32(999.0)
	for i, f := range morseFrequencies {
		diff := freq - f
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closestIdx = int32(i)
		}
	}
	return closestIdx
}

type MorseModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	state   *pb.MorseState
	pattern string
	timings []TimingEvent

	startTime time.Time
	lightOn   bool

	message     string
	messageType string
}

type TimingEvent struct {
	duration float64
	isOn     bool
}

func NewMorseModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *MorseModule {
	m := &MorseModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
		startTime: time.Now(),
	}

	m.state = mod.GetMorseState()
	if m.state != nil {
		m.pattern = m.state.GetDisplayedPattern()
		m.calculateTimings()
	}

	return m
}

func (m *MorseModule) calculateTimings() {
	if m.pattern == "" {
		m.timings = []TimingEvent{}
		return
	}

	m.timings = []TimingEvent{}

	for i := 0; i < len(m.pattern); i++ {
		char := m.pattern[i]

		if char == '.' {
			m.timings = append(m.timings, TimingEvent{duration: DOT_DURATION, isOn: true})
			m.timings = append(m.timings, TimingEvent{duration: SYMBOL_PAUSE, isOn: false})
		} else if char == '-' {
			m.timings = append(m.timings, TimingEvent{duration: DASH_DURATION, isOn: true})
			m.timings = append(m.timings, TimingEvent{duration: SYMBOL_PAUSE, isOn: false})
		} else if char == ' ' {
			m.timings = append(m.timings, TimingEvent{duration: LETTER_PAUSE - SYMBOL_PAUSE, isOn: false})
		}
	}
}

func (m *MorseModule) Init() tea.Cmd {
	return tea.Tick(TICK_INTERVAL, func(t time.Time) tea.Msg {
		return MorseTickMsg{Time: t}
	})
}

type MorseTickMsg struct {
	Time time.Time
}

func (m *MorseModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case MorseTickMsg:
		m.updateAnimation()
		return m, tea.Tick(TICK_INTERVAL, func(t time.Time) tea.Msg {
			return MorseTickMsg{Time: t}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			return m, m.changeFrequency(pb.IncrementDecrement_DECREMENT)
		case "right", "l":
			return m, m.changeFrequency(pb.IncrementDecrement_INCREMENT)
		case "enter":
			return m, m.transmit()
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

func (m *MorseModule) updateAnimation() {
	if len(m.timings) == 0 {
		m.lightOn = false
		return
	}

	totalDuration := 0.0
	for _, t := range m.timings {
		totalDuration += t.duration
	}

	if totalDuration == 0 {
		m.lightOn = false
		return
	}

	elapsed := time.Since(m.startTime).Seconds()
	normalizedTime := math.Mod(elapsed, totalDuration)

	runningTime := 0.0
	for _, t := range m.timings {
		if normalizedTime >= runningTime && normalizedTime < runningTime+t.duration {
			m.lightOn = t.isOn
			return
		}
		runningTime += t.duration
	}

	m.lightOn = false
}

func (m *MorseModule) changeFrequency(direction pb.IncrementDecrement) tea.Cmd {
	return func() tea.Msg {
		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_MorseInput{
				MorseInput: &pb.MorseInput{
					Input: &pb.MorseInput_FrequencyChange{
						FrequencyChange: &pb.MorseFrequencyChange{
							Direction: direction,
						},
					},
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if morseResult := result.GetMorseInputResult(); morseResult != nil {
			if morseState := morseResult.GetMorseState(); morseState != nil {
				oldPattern := m.pattern
				m.state = morseState
				m.pattern = morseState.GetDisplayedPattern()

				if m.pattern != "" && m.pattern != oldPattern {
					m.calculateTimings()
				}
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

func (m *MorseModule) transmit() tea.Cmd {
	return func() tea.Msg {
		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_MorseInput{
				MorseInput: &pb.MorseInput{
					Input: &pb.MorseInput_Tx{
						Tx: &pb.MorseTx{},
					},
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if result.GetStrike() {
			m.message = "STRIKE! Wrong frequency!"
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

func (m *MorseModule) View() string {
	frequency := float32(3.505)
	if m.state != nil {
		frequency = m.state.GetDisplayedFrequency()
	}

	idx := frequencyToIndex(frequency)

	light := m.renderLight()
	freq := m.renderFrequency(idx, frequency)
	tx := m.renderTXButton()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("MORSE CODE"),
		"",
		light,
		"",
		freq,
		"",
		tx,
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

func (m *MorseModule) renderLight() string {
	amberColor := lipgloss.Color("#FFB000")
	wireColor := lipgloss.Color("#8B4513")

	if m.lightOn {
		light := lipgloss.NewStyle().
			Background(amberColor).
			Foreground(wireColor).
			Render("●")
		return lipgloss.NewStyle().
			Foreground(wireColor).
			Render("━━━━━━━━━━━━━━━━━┤") + light + lipgloss.NewStyle().
			Foreground(wireColor).
			Render("├━━━━━━━━━━━━━━━━━")
	}

	light := lipgloss.NewStyle().
		Foreground(amberColor).
		Render("○")
	return lipgloss.NewStyle().
		Foreground(wireColor).
		Render("━━━━━━━━━━━━━━━━━┤") + light + lipgloss.NewStyle().
		Foreground(wireColor).
		Render("├━━━━━━━━━━━━━━━━━")
}

func (m *MorseModule) renderFrequency(idx int32, frequency float32) string {
	freqLabel := styles.Subtitle.Render("FREQUENCY")
	freqValue := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true).
		Render(fmt.Sprintf("%.3f MHz", frequency))

	arrows := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("◄──")

	arrowRight := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("──►")

	freqRow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		arrows,
		" ",
		freqValue,
		" ",
		arrowRight,
	)

	sliderBar := m.renderSlider(idx)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				freqLabel,
				"",
				freqRow,
				"",
				sliderBar,
			),
		)

	return box
}

func (m *MorseModule) renderSlider(idx int32) string {
	const totalPositions = 16
	sliderWidth := 24

	pos := int(float64(idx) / float64(totalPositions-1) * float64(sliderWidth-1))

	var sb strings.Builder
	for i := 0; i < sliderWidth; i++ {
		if i == pos {
			sb.WriteString("●")
		} else {
			sb.WriteString("─")
		}
	}

	sliderBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Render(sb.String())

	return sliderBar
}

func (m *MorseModule) renderTXButton() string {
	txLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#222222")).
		Bold(true).
		Render("  TX  ")

	button := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF6B6B")).
		Padding(0, 2).
		Render(txLabel)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Align(lipgloss.Center).
		Render(button)

	return box
}

func (m *MorseModule) ID() string {
	return m.mod.GetId()
}

func (m *MorseModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_MORSE
}

func (m *MorseModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *MorseModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}

	newState := mod.GetMorseState()
	if newState != nil {
		m.state = newState
		m.pattern = newState.GetDisplayedPattern()
		m.calculateTimings()
	}
}

func (m *MorseModule) Footer() string {
	return "[←/→] or [h/l] Adjust frequency | [ENTER] Transmit | [ESC] Back to bomb"
}
