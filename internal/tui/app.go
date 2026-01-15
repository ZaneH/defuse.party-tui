package tui

import (
	"context"
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"

	"github.com/ZaneH/defuse.party-tui/internal/client"
	"github.com/ZaneH/defuse.party-tui/internal/styles"
	"github.com/ZaneH/defuse.party-tui/internal/tui/modules"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
)

type Model struct {
	state AppState

	grpcAddr     string
	gameClient   client.GameClient
	sessionID    string
	bombs        []*pb.Bomb
	selectedBomb int

	currentFace int

	selectedModule int

	activeModule modules.ModuleModel
	moduleCache  map[string]modules.ModuleModel

	width  int
	height int
	err    error

	startedAt        time.Time
	duration         time.Duration
	strikeFlashUntil time.Time
	flashStrike      bool

	showQuitConfirm bool

	menuSelection     int
	sectionSelection  int
	missionSelection  int
	freePlaySelection int
	gameOverSelection int

	freePlayConfig    FreePlayConfig
	freePlayCursor    int
	freePlayInModules bool

	showManualDialog bool

	pendingGameConfig *pb.GameConfig
}

func NewProgramHandler(grpcAddr string) bubbletea.ProgramHandler {
	return func(sess ssh.Session) *tea.Program {
		_, _, active := sess.Pty()
		if active {
			lipgloss.SetColorProfile(termenv.ANSI256)
		}

		return tea.NewProgram(
			&Model{
				state:       StateMainMenu,
				grpcAddr:    grpcAddr,
				moduleCache: make(map[string]modules.ModuleModel),
			},
			tea.WithInput(sess),
			tea.WithOutput(sess),
			tea.WithAltScreen(),
		)
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) StartGame(config *pb.GameConfig) tea.Cmd {
	return func() tea.Msg {
		client, err := client.New(m.grpcAddr)
		if err != nil {
			return loadingErrorMsg{err: fmt.Errorf("failed to connect: %w", err)}
		}

		sessionID, err := client.CreateGame(context.Background(), config)
		if err != nil {
			client.Close()
			return loadingErrorMsg{err: fmt.Errorf("failed to create game: %w", err)}
		}

		bombs, err := client.GetBombs(context.Background(), sessionID)
		if err != nil {
			client.Close()
			return loadingErrorMsg{err: fmt.Errorf("failed to get bombs: %w", err)}
		}

		return gameReadyMsg{
			client:    client,
			sessionID: sessionID,
			bombs:     bombs,
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
		m.state = StateBombSelection
		m.gameClient = msg.client
		m.sessionID = msg.sessionID
		m.bombs = msg.bombs
		m.selectedBomb = 0
		m.currentFace = 0
		m.selectedModule = 0
		if len(msg.bombs) > 0 {
			bomb := msg.bombs[0]
			m.startedAt = time.Unix(int64(bomb.GetStartedAt()), 0)
			m.duration = time.Duration(bomb.GetTimerDuration()) * time.Second
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg{t: t}
		})

	case tickMsg:
		now := time.Now()

		if m.startedAt.IsZero() {
			return m, nil
		}

		elapsed := now.Sub(m.startedAt)
		remaining := m.duration - elapsed

		if remaining <= 0 {
			m.state = StateGameOver
			m.err = fmt.Errorf("time's up!")
			return m, tea.Quit
		}

		if m.flashStrike && now.After(m.strikeFlashUntil) {
			m.flashStrike = false
		}

		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg{t: t}
		})

	case modules.ModuleResultMsg:
		if msg.Err != nil {
			return m, nil
		}
		result := msg.Result
		if result.GetStrike() {
			m.flashStrike = true
			m.strikeFlashUntil = time.Now().Add(500 * time.Millisecond)
		}
		if bombStatus := result.GetBombStatus(); bombStatus != nil {
			bomb := m.getCurrentBomb()
			if bomb != nil {
				bomb.StrikeCount = bombStatus.GetStrikeCount()

				for _, cachedMod := range m.moduleCache {
					if cachedMod.ModuleType() == pb.Module_CLOCK {
						if clockMod, ok := cachedMod.(*modules.ClockModule); ok {
							clockMod.UpdateStrikes(bomb.StrikeCount)
						}
					}
				}
			}
		}
		if result.GetSolved() && m.activeModule != nil {
			m.activeModule.UpdateState(&pb.Module{
				Id:     result.GetModuleId(),
				Solved: true,
			})
		}
		if result.GetBombStatus().GetExploded() {
			m.state = StateGameOver
			m.err = fmt.Errorf("BOOM! The bomb exploded.")
			return m, tea.Quit
		}
		return m, nil

	case modules.BackToBombMsg:
		m.state = StateBombView
		m.activeModule = nil
		return m, nil

	case tea.KeyMsg:
		if m.showManualDialog {
			switch msg.String() {
			case "esc":
				m.showManualDialog = false
			}
			return m, nil
		}

		if m.showQuitConfirm {
			switch msg.String() {
			case "y", "Y":
				if m.gameClient != nil {
					m.gameClient.Close()
				}
				return m, tea.Quit
			case "n", "N", "esc":
				m.showQuitConfirm = false
				return m, nil
			}
			return m, nil
		}

		if cmd, handled := m.handleMenuKeys(msg.String()); handled {
			return m, cmd
		}

		switch m.state {
		case StateBombSelection:
			return m.handleBombSelectionKeys(msg)
		case StateBombView:
			return m.handleBombViewKeys(msg)
		}

		if msg.String() == "q" {
			m.showQuitConfirm = true
			return m, nil
		}
		if msg.String() == "ctrl+c" {
			if m.gameClient != nil {
				m.gameClient.Close()
			}
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.state == StateModuleActive && m.activeModule != nil {
		newModel, cmd := m.activeModule.Update(msg)
		if newModule, ok := newModel.(modules.ModuleModel); ok {
			m.activeModule = newModule
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleMenuKeys(key string) (tea.Cmd, bool) {
	switch m.state {
	case StateMainMenu:
		return m.handleMainMenuKeys(key)
	case StateSectionSelect:
		return m.handleSectionSelectKeys(key)
	case StateMissionSelect:
		return m.handleMissionSelectKeys(key)
	case StateFreePlayMenu:
		return m.handleFreePlayMenuKeys(key)
	case StateFreePlayAdvanced:
		return m.handleFreePlayAdvancedKeys(key)
	case StateGameOver:
		return m.handleGameOverKeys(key)
	}
	return nil, false
}

func (m *Model) handleGameOverKeys(key string) (tea.Cmd, bool) {
	handled := true
	switch key {
	case "up", "k":
		if m.gameOverSelection > 0 {
			m.gameOverSelection--
		}
	case "down", "j":
		if m.gameOverSelection < 1 {
			m.gameOverSelection++
		}
	case "enter":
		if m.gameOverSelection == 0 {
			m.resetToMainMenu()
		} else {
			if m.gameClient != nil {
				m.gameClient.Close()
			}
			return tea.Quit, true
		}
	default:
		handled = false
	}
	return nil, handled
}

func (m *Model) resetToMainMenu() {
	m.state = StateMainMenu
	m.menuSelection = 0
	m.gameClient = nil
	m.sessionID = ""
	m.bombs = nil
	m.selectedBomb = 0
	m.currentFace = 0
	m.selectedModule = 0
	m.activeModule = nil
	m.moduleCache = make(map[string]modules.ModuleModel)
	m.err = nil
	m.startedAt = time.Time{}
	m.duration = 0
	m.flashStrike = false
	m.showQuitConfirm = false
	m.showManualDialog = false
	m.pendingGameConfig = nil
}

func (m *Model) handleBombSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.state = StateBombView
		m.selectedModule = 0
		m.currentFace = 0
		return m, nil
	case "up", "k":
		if m.selectedBomb > 0 {
			m.selectedBomb--
		}
	case "down", "j":
		if m.selectedBomb < len(m.bombs)-1 {
			m.selectedBomb++
		}
	case "q":
		m.showQuitConfirm = true
		return m, nil
	case "ctrl+c":
		if m.gameClient != nil {
			m.gameClient.Close()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleBombViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	faceModules := m.getCurrentFaceModules()

	switch msg.String() {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		moduleIdx := int(msg.String()[0] - '1')
		if moduleIdx >= 0 && moduleIdx < len(faceModules) {
			m.state = StateModuleActive
			mod := faceModules[moduleIdx]
			moduleID := mod.GetId()

			if cached, exists := m.moduleCache[moduleID]; exists {
				m.activeModule = cached
				return m, nil
			} else {
				if mod.GetType() == pb.Module_CLOCK {
					m.activeModule = m.createClockModule(mod)
				} else {
					m.activeModule = modules.NewModule(mod, m.gameClient, m.sessionID, m.getCurrentBomb().GetId())
				}
				m.moduleCache[moduleID] = m.activeModule
				return m, m.activeModule.Init()
			}
		}
	case "enter":
		if len(faceModules) > 0 {
			m.state = StateModuleActive
			mod := faceModules[m.selectedModule]
			moduleID := mod.GetId()

			if cached, exists := m.moduleCache[moduleID]; exists {
				m.activeModule = cached
				return m, nil
			} else {
				if mod.GetType() == pb.Module_CLOCK {
					m.activeModule = m.createClockModule(mod)
				} else {
					m.activeModule = modules.NewModule(mod, m.gameClient, m.sessionID, m.getCurrentBomb().GetId())
				}
				m.moduleCache[moduleID] = m.activeModule
				return m, m.activeModule.Init()
			}
		}
	case "esc", "tab", "b":
		m.state = StateBombSelection
		m.moduleCache = make(map[string]modules.ModuleModel)
		return m, nil
	case "<":
		if m.currentFace > 0 {
			m.currentFace--
			m.selectedModule = 0
		}
	case ">":
		maxFace := m.maxFaceIndex()
		if m.currentFace < maxFace {
			m.currentFace++
			m.selectedModule = 0
		}
	case "up", "k":
		if m.selectedModule > 0 {
			m.selectedModule--
		}
	case "down", "j":
		if m.selectedModule < len(faceModules)-1 {
			m.selectedModule++
		}
	case "left", "h":
		if m.selectedModule > 0 {
			m.selectedModule--
		}
	case "right", "l":
		if m.selectedModule < len(faceModules)-1 {
			m.selectedModule++
		}
	case "q":
		m.showQuitConfirm = true
		return m, nil
	case "ctrl+c":
		if m.gameClient != nil {
			m.gameClient.Close()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) getCurrentBomb() *pb.Bomb {
	if m.selectedBomb >= 0 && m.selectedBomb < len(m.bombs) {
		return m.bombs[m.selectedBomb]
	}
	return nil
}

func (m *Model) createClockModule(mod *pb.Module) modules.ModuleModel {
	bomb := m.getCurrentBomb()
	if bomb == nil {
		return modules.NewUnimplementedModule(mod)
	}
	return modules.NewClockModule(
		mod,
		m.startedAt,
		m.duration,
		bomb.GetStrikeCount(),
		bomb.GetMaxStrikes(),
	)
}

func (m *Model) getCurrentFaceModules() []*pb.Module {
	bomb := m.getCurrentBomb()
	if bomb == nil {
		return nil
	}

	var faceModules []*pb.Module
	for _, mod := range bomb.GetModules() {
		if mod.GetPosition() != nil && mod.GetPosition().GetFace() == int32(m.currentFace) {
			faceModules = append(faceModules, mod)
		}
	}

	for _, mod := range bomb.GetModules() {
		if mod.GetPosition() == nil {
			faceModules = append(faceModules, mod)
		}
	}

	sort.Slice(faceModules, func(i, j int) bool {
		posI := faceModules[i].GetPosition()
		posJ := faceModules[j].GetPosition()

		if posI == nil && posJ == nil {
			return faceModules[i].GetId() < faceModules[j].GetId()
		}
		if posI == nil {
			return false
		}
		if posJ == nil {
			return true
		}

		if posI.GetRow() != posJ.GetRow() {
			return posI.GetRow() < posJ.GetRow()
		}
		return posI.GetCol() < posJ.GetCol()
	})

	return faceModules
}

func (m *Model) maxFaceIndex() int {
	bomb := m.getCurrentBomb()
	if bomb == nil {
		return 0
	}

	maxFace := 0
	for _, mod := range bomb.GetModules() {
		if mod.GetPosition() != nil {
			if face := int(mod.GetPosition().GetFace()); face > maxFace {
				maxFace = face
			}
		}
	}
	return maxFace
}

func (m *Model) View() string {
	var view string

	switch m.state {
	case StateMainMenu:
		view = m.mainMenuView()
	case StateSectionSelect:
		view = m.sectionSelectView()
	case StateMissionSelect:
		view = m.missionSelectView()
	case StateFreePlayMenu:
		view = m.freePlayMenuView()
	case StateFreePlayAdvanced:
		view = m.freePlayAdvancedView()
	case StateLoading:
		view = m.loadingView()
	case StateGameOver:
		view = m.gameOverView()
	case StateBombSelection:
		view = m.bombSelectionView()
	case StateBombView:
		view = m.bombView()
	case StateModuleActive:
		if m.activeModule != nil {
			header := m.renderHeader(time.Now())
			footer := m.renderFooter()
			content := m.activeModule.View()
			view = lipgloss.JoinVertical(
				lipgloss.Top,
				header,
				styles.ContentBox.Render(content),
				footer,
			)
		} else {
			view = m.errorView()
		}
	}

	if m.showManualDialog {
		dialog := styles.DialogBox.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styles.Title.Render("MANUAL"),
				"",
				styles.Subtitle.Render("Open this URL in your browser:"),
				"",
				styles.Active.Render(hyperlink("https://bombmanual.com/", "https://bombmanual.com/")),
				"",
				styles.Help.Render("(Click the link or copy to your browser)"),
				"",
				styles.Help.Render("[ESC] Back"),
			),
		)
		view = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialog,
		)
	}

	if m.showQuitConfirm {
		dialog := styles.DialogBox.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styles.Warning.Bold(true).Render("Quit game?"),
				"",
				styles.Help.Render("[Y] Yes  [N] No"),
			),
		)
		view = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialog,
		)
	}

	return view
}

func (m *Model) loadingView() string {
	return styles.Center(
		lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES"),
			"",
			styles.Subtitle.Render("Creating game..."),
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
			styles.Title.Render("GAME OVER"),
			"",
			styles.Error.Render(errMsg),
		),
		m.width, m.height,
	)
}
