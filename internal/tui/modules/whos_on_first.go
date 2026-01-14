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

type WhosOnFirstModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	message     string
	messageType string
}

func NewWhosOnFirstModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *WhosOnFirstModule {
	return &WhosOnFirstModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
	}
}

func (m *WhosOnFirstModule) Init() tea.Cmd {
	return nil
}

func (m *WhosOnFirstModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5", "6":
			pos := int(msg.String()[0] - '1')
			return m, m.pressButton(pos)
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

func (m *WhosOnFirstModule) pressButton(pos int) tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetWhosOnFirstState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no whos on first state")}
		}

		buttonWords := state.GetButtonWords()
		if pos < 0 || pos >= len(buttonWords) {
			return ModuleResultMsg{Err: fmt.Errorf("invalid position")}
		}

		word := buttonWords[pos]

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_WhosOnFirstInput{
				WhosOnFirstInput: &pb.WhosOnFirstInput{
					Word: word,
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if wofResult := result.GetWhosOnFirstInputResult(); wofResult != nil {
			if wofState := wofResult.GetWhosOnFirstState(); wofState != nil {
				m.mod.State = &pb.Module_WhosOnFirstState{WhosOnFirstState: wofState}
			}
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

func (m *WhosOnFirstModule) View() string {
	state := m.mod.GetWhosOnFirstState()
	if state == nil {
		return styles.Error.Render("No Who's On First state available")
	}

	screenWord := state.GetScreenWord()
	buttonWords := state.GetButtonWords()
	stage := int(state.GetStage())

	screenStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#FFD93D")).
		Padding(1, 2).
		Width(14).
		Height(3)

	screenContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render(screenWord),
	)

	screenSection := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(screenStyle.Render(screenContent))

	var buttonRows []string
	for row := 0; row < 3; row++ {
		leftPos := row * 2
		rightPos := row*2 + 1

		leftButton := renderWhosOnFirstButton(buttonWords[leftPos], leftPos+1)
		rightButton := renderWhosOnFirstButton(buttonWords[rightPos], rightPos+1)

		rowContent := lipgloss.JoinHorizontal(
			lipgloss.Center,
			leftButton,
			"  ",
			rightButton,
		)
		buttonRows = append(buttonRows, rowContent)
	}

	buttonSection := lipgloss.JoinVertical(
		lipgloss.Center,
		buttonRows...,
	)

	var stageIndicators []string
	for i := 1; i <= 5; i++ {
		if i < stage {
			stageIndicators = append(stageIndicators, styles.Success.Render("●"))
		} else {
			stageIndicators = append(stageIndicators, styles.Help.Render("○"))
		}
	}

	stageRow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().Render(" "),
		lipgloss.JoinHorizontal(lipgloss.Center, stageIndicators...),
		lipgloss.NewStyle().Render(" "),
	)

	mainContent := lipgloss.JoinVertical(
		lipgloss.Center,
		screenSection,
		"",
		buttonSection,
		"",
		stageRow,
	)

	content := lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styles.Title.Render("WHO'S ON FIRST"),
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

func renderWhosOnFirstButton(word string, position int) string {
	buttonStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(10).
		Height(3)

	buttonContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Help.Render(fmt.Sprintf("[%d]", position)),
		lipgloss.NewStyle().Bold(true).Render(word),
	)

	return buttonStyle.Render(buttonContent)
}

func (m *WhosOnFirstModule) ID() string {
	return m.mod.GetId()
}

func (m *WhosOnFirstModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_WHOS_ON_FIRST
}

func (m *WhosOnFirstModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *WhosOnFirstModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}
}

func (m *WhosOnFirstModule) Footer() string {
	return "[1-6] Select word | [ESC] Back to bomb"
}
