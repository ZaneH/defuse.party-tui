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

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	"github.com/ZaneH/keep-talking-tui/internal/tui/modules"
	pb "github.com/ZaneH/keep-talking/pkg/proto"
)

type AppState int

const (
	StateLoading AppState = iota
	StateBombSelection
	StateBombView
	StateModuleActive
	StateGameOver
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
}

func NewProgramHandler(grpcAddr string) bubbletea.ProgramHandler {
	return func(sess ssh.Session) *tea.Program {
		_, _, active := sess.Pty()
		if active {
			lipgloss.SetColorProfile(termenv.ANSI256)
		}

		return tea.NewProgram(
			&Model{
				state:       StateLoading,
				grpcAddr:    grpcAddr,
				moduleCache: make(map[string]modules.ModuleModel),
			},
			tea.WithInput(sess),
			tea.WithOutput(sess),
			tea.WithAltScreen(),
		)
	}
}

type loadingErrorMsg struct{ err error }
type gameReadyMsg struct {
	client    client.GameClient
	sessionID string
	bombs     []*pb.Bomb
}
type tickMsg struct{ t time.Time }

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		client, err := client.New(m.grpcAddr)
		if err != nil {
			return loadingErrorMsg{err: fmt.Errorf("failed to connect: %w", err)}
		}

		sessionID, err := client.CreateGame(context.Background())
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

				// Update cached clock module strikes
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
		// Handle quit confirmation dialog
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
			return m, nil // Ignore other keys while dialog is shown
		}

		switch m.state {
		case StateBombSelection:
			return m.handleBombSelectionKeys(msg)
		case StateBombView:
			return m.handleBombViewKeys(msg)
		}

		// Fallback quit handler for states not explicitly handled above
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
				// Special handling for clock module
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
	case StateLoading:
		view = m.loadingView()
	case StateGameOver:
		view = m.errorView()
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

	// Overlay quit confirmation dialog
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
			styles.Title.Render("GAME OVER"),
			"",
			styles.Error.Render(errMsg),
		),
		m.width, m.height,
	)
}

func (m *Model) bombSelectionView() string {
	header := m.renderHeader(time.Now())
	footer := m.renderFooter()

	var bombList []string
	for i, bomb := range m.bombs {
		serial := bomb.GetSerialNumber()
		numModules := len(bomb.GetModules())
		if i == m.selectedBomb {
			bombList = append(bombList, styles.Active.Render(fmt.Sprintf("> BOMB %d: Serial %s  [%d modules]", i+1, serial, numModules)))
		} else {
			bombList = append(bombList, fmt.Sprintf("  BOMB %d: Serial %s  [%d modules]", i+1, serial, numModules))
		}
	}

	if len(bombList) == 0 {
		bombList = append(bombList, "  No bombs available")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("SELECT A BOMB"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, bombList...),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		styles.ContentBox.Render(content),
		footer,
	)
}

func (m *Model) bombView() string {
	header := m.renderHeader(time.Now())
	footer := m.renderFooter()

	faceModules := m.getCurrentFaceModules()

	var modules []string
	for i, mod := range faceModules {
		modTypeName := m.moduleTypeName(mod.GetType())
		timer := m.getModuleTimer(mod)
		solved := mod.GetSolved()
		status := "○ PENDING"
		if solved {
			status = "✓ SOLVED"
		}

		// Format module line with right-aligned timer
		moduleLine := fmt.Sprintf("[%d] %s", i+1, modTypeName)
		if timer != "" {
			// Right-align timer within ~40 char width
			nameLen := len(moduleLine)
			timerLen := len(timer)
			padding := 40 - nameLen - timerLen
			if padding < 1 {
				padding = 1
			}
			moduleLine = fmt.Sprintf("%s%*s%s", moduleLine, padding, "", timer)
		}

		if i == m.selectedModule {
			modules = append(modules, styles.Active.Render(moduleLine))
			modules = append(modules, styles.Active.Render(fmt.Sprintf("  > SELECTED    %s", status)))
		} else {
			modules = append(modules, moduleLine)
			modules = append(modules, fmt.Sprintf("  %s", status))
		}
		modules = append(modules, "")
	}

	faceName := "FRONT"
	if m.currentFace == 1 {
		faceName = "BACK"
	} else if m.currentFace > 1 {
		faceName = fmt.Sprintf("FACE %d", m.currentFace+1)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render(fmt.Sprintf("BOMB %d - %s", m.selectedBomb+1, faceName)),
		"",
		lipgloss.JoinVertical(lipgloss.Left, modules...),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		styles.ContentBox.Render(content),
		footer,
	)
}

func (m *Model) moduleTypeName(t pb.Module_ModuleType) string {
	switch t {
	case pb.Module_CLOCK:
		return "CLOCK"
	case pb.Module_WIRES:
		return "WIRES"
	case pb.Module_PASSWORD:
		return "PASSWORD"
	case pb.Module_BIG_BUTTON:
		return "BIG BUTTON"
	case pb.Module_SIMON:
		return "SIMON"
	case pb.Module_KEYPAD:
		return "KEYPAD"
	case pb.Module_WHOS_ON_FIRST:
		return "WHO'S ON FIRST"
	case pb.Module_MEMORY:
		return "MEMORY"
	case pb.Module_MORSE:
		return "MORSE CODE"
	case pb.Module_NEEDY_VENT_GAS:
		return "VENT GAS"
	case pb.Module_NEEDY_KNOB:
		return "KNOB"
	case pb.Module_MAZE:
		return "MAZE"
	default:
		return "UNKNOWN"
	}
}

func (m *Model) getModuleTimer(mod *pb.Module) string {
	now := time.Now()

	if mod.GetType() == pb.Module_CLOCK {
		elapsed := now.Sub(m.startedAt)
		remaining := m.duration - elapsed
		if remaining < 0 {
			remaining = 0
		}
		mins := int(remaining.Minutes())
		secs := int(remaining.Seconds()) % 60
		return fmt.Sprintf("[%d:%02d]", mins, secs)
	}

	if mod.GetType() == pb.Module_NEEDY_VENT_GAS {
		if state := mod.GetNeedyVentGasState(); state != nil {
			if state.GetCountdownStartedAt() == 0 {
				return "" // Inactive, no timer
			}
			startedAt := time.Unix(state.GetCountdownStartedAt(), 0)
			elapsed := now.Sub(startedAt)
			remaining := time.Duration(state.GetCountdownDuration())*time.Second - elapsed
			if remaining < 0 {
				remaining = 0
			}
			secs := int(remaining.Seconds())
			return fmt.Sprintf("[0:%02d]", secs)
		}
	}

	if mod.GetType() == pb.Module_NEEDY_KNOB {
		if state := mod.GetNeedyKnobState(); state != nil {
			if state.GetCountdownStartedAt() == 0 {
				return "" // Inactive, no timer
			}
			startedAt := time.Unix(state.GetCountdownStartedAt(), 0)
			elapsed := now.Sub(startedAt)
			remaining := time.Duration(state.GetCountdownDuration())*time.Second - elapsed
			if remaining < 0 {
				remaining = 0
			}
			secs := int(remaining.Seconds())
			return fmt.Sprintf("[0:%02d]", secs)
		}
	}

	return ""
}

func (m *Model) renderHeader(now time.Time) string {
	bomb := m.getCurrentBomb()
	if bomb == nil {
		return styles.HeaderBox.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES - DEFUSER TERMINAL"),
			),
		)
	}

	elapsed := now.Sub(m.startedAt)
	remaining := m.duration - elapsed
	if remaining < 0 {
		remaining = 0
	}

	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	timerStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	timerStyle := styles.Normal
	if remaining < 30*time.Second {
		timerStyle = styles.Error
	} else if remaining < 60*time.Second {
		timerStyle = styles.Warning
	}

	if m.flashStrike {
		timerStyle = styles.Strike
	}

	strikes := ""
	for i := int32(0); i < bomb.GetMaxStrikes(); i++ {
		if i < bomb.GetStrikeCount() {
			strikes += "[X] "
		} else {
			strikes += "[ ] "
		}
	}

	serial := bomb.GetSerialNumber()
	batteries := bomb.GetBatteries()
	ports := bomb.GetPorts()
	portStr := ""
	if len(ports) > 0 {
		var portNames []string
		for _, p := range ports {
			switch p {
			case pb.Port_DVID:
				portNames = append(portNames, "DVI")
			case pb.Port_RCA:
				portNames = append(portNames, "RCA")
			case pb.Port_PS2:
				portNames = append(portNames, "PS2")
			case pb.Port_RJ45:
				portNames = append(portNames, "RJ45")
			case pb.Port_SERIAL:
				portNames = append(portNames, "SER")
			}
		}
		portStr = fmt.Sprintf("Ports: %s", joinStrings(portNames, ", "))
	}

	batteryStr := ""
	if batteries > 0 {
		batteryStr = fmt.Sprintf("Batteries: %d", batteries)
	}

	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES"),
		"  ",
		timerStyle.Render(fmt.Sprintf("Time: %s", timerStr)),
		"  ",
		styles.Normal.Render(fmt.Sprintf("Serial: %s", serial)),
	)

	if batteryStr != "" || portStr != "" {
		headerContent = lipgloss.JoinVertical(
			lipgloss.Left,
			headerContent,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				styles.Normal.Render(strikes),
				"  ",
				styles.Help.Render(batteryStr),
				"  ",
				styles.Help.Render(portStr),
			),
		)
	} else {
		headerContent = lipgloss.JoinVertical(
			lipgloss.Left,
			headerContent,
			styles.Normal.Render(strikes),
		)
	}

	return styles.HeaderBox.Render(headerContent)
}

func (m *Model) renderFooter() string {
	hint := ""
	switch m.state {
	case StateBombSelection:
		hint = "[ENTER] Pick up bomb | [↑/↓] Navigate | [Q]uit"
	case StateBombView:
		hint = "[1-9] Select module | [<]/[>] Flip face | [ESC] Put down | [Q]uit"
	case StateModuleActive:
		hint = m.activeModule.Footer()
	}
	return styles.FooterBox.Render(styles.Help.Render(hint))
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
