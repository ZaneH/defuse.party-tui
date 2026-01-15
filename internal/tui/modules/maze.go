package modules

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/defuse.party-tui/internal/client"
	"github.com/ZaneH/defuse.party-tui/internal/styles"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
)

type MazeModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	playerX  int
	playerY  int
	goalX    int
	goalY    int
	marker1X int
	marker1Y int
	marker2X int
	marker2Y int

	message     string
	messageType string
}

func NewMazeModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *MazeModule {
	m := &MazeModule{
		mod:       mod,
		client:    client,
		sessionID: sessionID,
		bombID:    bombID,
	}

	state := mod.GetMazeState()
	if state != nil {
		m.marker1X = int(state.GetMarker_1().GetX())
		m.marker1Y = int(state.GetMarker_1().GetY())
		m.marker2X = int(state.GetMarker_2().GetX())
		m.marker2Y = int(state.GetMarker_2().GetY())
	}

	m.playerX = int(state.GetPlayerPosition().GetX())
	m.playerY = int(state.GetPlayerPosition().GetY())
	m.goalX = int(state.GetGoalPosition().GetX())
	m.goalY = int(state.GetGoalPosition().GetY())

	return m
}

func (m *MazeModule) updatePositionsFromState() {
	state := m.mod.GetMazeState()
	if state == nil {
		return
	}

	m.playerX = int(state.GetPlayerPosition().GetX())
	m.playerY = int(state.GetPlayerPosition().GetY())
	m.goalX = int(state.GetGoalPosition().GetX())
	m.goalY = int(state.GetGoalPosition().GetY())
}

func (m *MazeModule) Init() tea.Cmd {
	return nil
}

func (m *MazeModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "w":
			return m, m.move(pb.CardinalDirection_NORTH)
		case "down", "j", "s":
			return m, m.move(pb.CardinalDirection_SOUTH)
		case "left", "h", "a":
			return m, m.move(pb.CardinalDirection_WEST)
		case "right", "l", "d":
			return m, m.move(pb.CardinalDirection_EAST)
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

func (m *MazeModule) move(direction pb.CardinalDirection) tea.Cmd {
	return func() tea.Msg {
		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input: &pb.PlayerInput_MazeInput{
				MazeInput: &pb.MazeInput{
					Direction: direction,
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
		} else if result.GetStrike() {
			m.message = "STRIKE! Invalid move!"
			m.messageType = "error"
		}

		if mazeResult := result.GetMazeInputResult(); mazeResult != nil {
			if mazeState := mazeResult.GetMazeState(); mazeState != nil {
				m.mod.State = &pb.Module_MazeState{MazeState: mazeState}
				m.updatePositionsFromState()
			}
		}

		return ModuleResultMsg{Result: result}
	}
}

func (m *MazeModule) View() string {
	grid := m.renderGrid()

	content := lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styles.Title.Render("MAZE"),
				"",
				grid,
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

func (m *MazeModule) renderGrid() string {
	var rows []string

	for y := 0; y <= 5; y++ {
		var cells []string
		for x := 0; x <= 5; x++ {
			cells = append(cells, m.renderCell(x, y))
		}
		row := lipgloss.JoinHorizontal(lipgloss.Center, cells...)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Padding(0, 2).Render(lipgloss.JoinVertical(lipgloss.Center, rows...)),
	)
}

func (m *MazeModule) renderCell(x, y int) string {
	cellStyle := lipgloss.NewStyle().
		Width(3).
		Height(1).
		Align(lipgloss.Center)

	switch {
	case x == m.playerX && y == m.playerY:
		return cellStyle.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#4DABF7")).Render("●")
	case x == m.goalX && y == m.goalY:
		return cellStyle.Foreground(lipgloss.Color("#FF6B6B")).Render("▲")
	case (x == m.marker1X && y == m.marker1Y) || (x == m.marker2X && y == m.marker2Y):
		return cellStyle.Foreground(lipgloss.Color("#6BCB77")).Render("◎")
	default:
		return cellStyle.Foreground(lipgloss.Color("#666666")).Render("○")
	}
}

func (m *MazeModule) ID() string {
	return m.mod.GetId()
}

func (m *MazeModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_MAZE
}

func (m *MazeModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *MazeModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}
}

func (m *MazeModule) Footer() string {
	return "[↑/↓/←/→] or [W/A/S/D] or [H/J/K/L] Move | [ESC] Back to bomb"
}
