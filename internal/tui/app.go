package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
)

type AppState int

const (
	StateLoading AppState = iota
	StateModuleList
	StateModuleActive
	StateGameOver
)

type Model struct {
	state      AppState
	grpcAddr   string
	gameClient client.GameClient
	sessionID  string
	width      int
	height     int
	err        error
	message    string
}

func NewProgramHandler(grpcAddr string) bubbletea.ProgramHandler {
	return func(sess ssh.Session) *tea.Program {
		return tea.NewProgram(
			&Model{
				state:    StateLoading,
				grpcAddr: grpcAddr,
			},
			tea.WithInput(sess),
			tea.WithOutput(sess),
			tea.WithAltScreen(),
		)
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		client, err := client.New(m.grpcAddr)
		if err != nil {
			return loadingErrorMsg{err: err}
		}

		sessionID, err := client.CreateGame(context.Background())
		if err != nil {
			client.Close()
			return loadingErrorMsg{err: fmt.Errorf("failed to create game: %w", err)}
		}

		return gameReadyMsg{
			client:    client,
			sessionID: sessionID,
		}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadingErrorMsg:
		m.state = StateGameOver
		m.err = msg.err
		return m, tea.Quit

	case gameReadyMsg:
		m.state = StateModuleList
		m.gameClient = msg.client
		m.sessionID = msg.sessionID
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.gameClient != nil {
				m.gameClient.Close()
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Model) View() string {
	switch m.state {
	case StateLoading:
		return m.loadingView()
	case StateGameOver:
		return m.errorView()
	default:
		return m.moduleListView()
	}
}

func (m *Model) loadingView() string {
	return styles.Center(
		lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES"),
			"",
			styles.Subtitle.Render("Connecting to game server..."),
		),
		m.width, m.height,
	)
}

func (m *Model) errorView() string {
	errMsg := "Unknown error"
	if m.err != nil {
		errMsg = m.err.Error()
	}
	return styles.Center(
		lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Title.Render("ERROR"),
			"",
			styles.Error.Render(errMsg),
		),
		m.width, m.height,
	)
}

func (m *Model) moduleListView() string {
	header := m.renderHeader(time.Now())
	footer := m.renderFooter()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("MODULE SELECTION"),
		"",
		styles.Help.Render("Press [ENTER] to select a module, [Q] to quit"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		styles.ContentBox.Render(content),
		footer,
	)
}

func (m *Model) renderHeader(now time.Time) string {
	timerStyle := styles.Normal
	if now.Second()%2 == 0 {
		timerStyle = styles.Warning
	}

	return styles.HeaderBox.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES - DEFUSER TERMINAL"),
			"  ",
			timerStyle.Render(fmt.Sprintf("Time: --:--")),
			"  ",
			styles.Normal.Render("Strikes: [ ] [ ] [ ]"),
		),
	)
}

func (m *Model) renderFooter() string {
	return styles.FooterBox.Render(
		styles.Help.Render("Commands: [ENTER] Select | [Q]uit"),
	)
}

type loadingErrorMsg struct{ err error }
type gameReadyMsg struct {
	client    client.GameClient
	sessionID string
}
