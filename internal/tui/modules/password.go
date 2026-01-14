package modules

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking-tui/proto"
)

type PasswordModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	selectedColumn int

	message     string
	messageType string
}

func NewPasswordModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *PasswordModule {
	return &PasswordModule{
		mod:            mod,
		client:         client,
		sessionID:      sessionID,
		bombID:         bombID,
		selectedColumn: 0,
	}
}

func (m *PasswordModule) Init() tea.Cmd {
	return nil
}

func (m *PasswordModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5":
			m.selectedColumn = int(msg.String()[0] - '1')
			return m, nil
		case "up", "k":
			return m, m.changeLetter(m.selectedColumn, pb.IncrementDecrement_INCREMENT)
		case "down", "j":
			return m, m.changeLetter(m.selectedColumn, pb.IncrementDecrement_DECREMENT)
		case "enter":
			return m, m.submit()
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

func (m *PasswordModule) changeLetter(col int, direction pb.IncrementDecrement) tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetPasswordState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no password state")}
		}

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_PasswordInput{
				PasswordInput: &pb.PasswordInput{
					Input: &pb.PasswordInput_LetterChange{
						LetterChange: &pb.LetterChange{
							LetterIndex: int32(col),
							Direction:   direction,
						},
					},
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if pwdResult := result.GetPasswordInputResult(); pwdResult != nil {
			if pwdState := pwdResult.GetPasswordState(); pwdState != nil {
				m.mod.State = &pb.Module_PasswordState{PasswordState: pwdState}
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

func (m *PasswordModule) submit() tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetPasswordState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no password state")}
		}

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_PasswordInput{
				PasswordInput: &pb.PasswordInput{
					Input: &pb.PasswordInput_Submit{
						Submit: &pb.PasswordSubmit{},
					},
				},
			},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		if result.GetStrike() {
			m.message = "STRIKE! Wrong password!"
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

func (m *PasswordModule) View() string {
	state := m.mod.GetPasswordState()
	if state == nil {
		return styles.Error.Render("No password state available")
	}

	letters := state.GetLetters()
	if len(letters) == 0 {
		return styles.Subtitle.Render("No letters displayed")
	}

	var boxes []string
	for i := 0; i < len(letters); i++ {
		letter := strings.ToUpper(string(letters[i]))

		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Width(5).
			Height(5)

		if i == m.selectedColumn {
			boxStyle = boxStyle.BorderForeground(lipgloss.Color("#4ECDC4"))
		}

		boxContent := lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Help.Render(fmt.Sprintf("[%d]", i+1)),
			styles.Help.Render(" ↑ "),
			lipgloss.NewStyle().Bold(true).Render(letter),
			styles.Help.Render(" ↓ "),
		)

		boxes = append(boxes, boxStyle.Render(boxContent))
	}

	boxesRow := lipgloss.JoinHorizontal(lipgloss.Center, boxes...)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("PASSWORD"),
		"",
		boxesRow,
		"",
		styles.Subtitle.Render("Press [ENTER] to submit"),
	)

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

func (m *PasswordModule) ID() string {
	return m.mod.GetId()
}

func (m *PasswordModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_PASSWORD
}

func (m *PasswordModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *PasswordModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}
}

func (m *PasswordModule) Footer() string {
	return "[1-5] Select column | [↑/↓] Change letter | [ENTER] Submit | [ESC] Back to bomb"
}
