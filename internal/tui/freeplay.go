package tui

import (
	"github.com/ZaneH/defuse.party-tui/internal/styles"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var freePlayPresets = []string{
	"EASY",
	"MEDIUM",
	"HARD",
	"EXPERT",
	"ADVANCED...",
}

var freePlayModuleTypes = []pb.Module_ModuleType{
	pb.Module_WIRES,
	pb.Module_PASSWORD,
	pb.Module_BIG_BUTTON,
	pb.Module_SIMON,
	pb.Module_KEYPAD,
	pb.Module_WHOS_ON_FIRST,
	pb.Module_MEMORY,
	pb.Module_MORSE,
	pb.Module_MAZE,
	pb.Module_NEEDY_VENT_GAS,
	pb.Module_NEEDY_KNOB,
}

var freePlayModuleNames = []string{
	"Wires",
	"Password",
	"Big Button",
	"Simon",
	"Keypad",
	"Who's On First",
	"Memory",
	"Morse Code",
	"Maze",
	"Needy Vent",
	"Needy Knob",
}

func (m *Model) freePlayMenuView() string {
	var items []string
	for i, preset := range freePlayPresets {
		if i == m.freePlaySelection {
			items = append(items, styles.Active.Render("> "+preset))
		} else {
			items = append(items, "  "+preset)
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("FREE PLAY"),
		"",
		styles.Subtitle.Render("Choose a difficulty or customize your own"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, items...),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		styles.HeaderBox.Render(styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES")),
		styles.ContentBox.Render(content),
		m.renderFooter(),
	)
}

func (m *Model) handleFreePlayMenuKeys(key string) (tea.Cmd, bool) {
	handled := true
	switch key {
	case "up", "k":
		if m.freePlaySelection > 0 {
			m.freePlaySelection--
		}
	case "down", "j":
		if m.freePlaySelection < len(freePlayPresets)-1 {
			m.freePlaySelection++
		}
	case "enter":
		switch m.freePlaySelection {
		case 0:
			m.pendingGameConfig = &pb.GameConfig{
				ConfigType: &pb.GameConfig_Level{
					Level: &pb.LevelConfig{Level: 1},
				},
			}
			m.state = StateLoading
			return m.StartGame(m.pendingGameConfig), true
		case 1:
			m.pendingGameConfig = &pb.GameConfig{
				ConfigType: &pb.GameConfig_Level{
					Level: &pb.LevelConfig{Level: 3},
				},
			}
			m.state = StateLoading
			return m.StartGame(m.pendingGameConfig), true
		case 2:
			m.pendingGameConfig = &pb.GameConfig{
				ConfigType: &pb.GameConfig_Level{
					Level: &pb.LevelConfig{Level: 5},
				},
			}
			m.state = StateLoading
			return m.StartGame(m.pendingGameConfig), true
		case 3:
			m.pendingGameConfig = &pb.GameConfig{
				ConfigType: &pb.GameConfig_Level{
					Level: &pb.LevelConfig{Level: 7},
				},
			}
			m.state = StateLoading
			return m.StartGame(m.pendingGameConfig), true
		case 4:
			m.state = StateFreePlayAdvanced
			m.freePlayConfig = DefaultFreePlayConfig()
			m.freePlayCursor = 0
			m.freePlayInModules = false
		}
	case "esc":
		m.state = StateMainMenu
		m.menuSelection = 0
	default:
		handled = false
	}
	return nil, handled
}

func (m *Model) freePlayAdvancedView() string {
	timerDisplay := formatTimeString(m.freePlayConfig.TimerSeconds)

	var rows []string
	rows = append(rows, styles.Title.Render("FREE PLAY - ADVANCED"))

	timerRow := "  Timer:           "
	if m.freePlayCursor == 0 {
		timerRow += styles.Active.Render("◀ " + timerDisplay + " ▶")
	} else {
		timerRow += "◀ " + timerDisplay + " ▶"
	}
	rows = append(rows, timerRow)

	strikesRow := "  Max Strikes:     "
	if m.freePlayCursor == 1 {
		strikesRow += styles.Active.Render("◀ " + formatTwoDigits(m.freePlayConfig.MaxStrikes) + " ▶")
	} else {
		strikesRow += "◀ " + formatTwoDigits(m.freePlayConfig.MaxStrikes) + " ▶"
	}
	rows = append(rows, strikesRow)

	facesRow := "  Bomb Faces:      "
	if m.freePlayCursor == 2 {
		facesRow += styles.Active.Render("◀ " + formatTwoDigits(m.freePlayConfig.NumFaces) + " ▶")
	} else {
		facesRow += "◀ " + formatTwoDigits(m.freePlayConfig.NumFaces) + " ▶"
	}
	rows = append(rows, facesRow)

	modulesRow := "  Modules/Face:    "
	if m.freePlayCursor == 3 {
		modulesRow += styles.Active.Render("◀ " + formatTwoDigits(m.freePlayConfig.ModulesPerFace) + " ▶")
	} else {
		modulesRow += "◀ " + formatTwoDigits(m.freePlayConfig.ModulesPerFace) + " ▶"
	}
	rows = append(rows, modulesRow)

	rows = append(rows, "")
	rows = append(rows, "  Modules:")
	rows = append(rows, "")

	moduleCols := [][]string{}
	for i := 0; i < len(freePlayModuleTypes); i += 2 {
		var row []string
		for j := 0; j < 2 && i+j < len(freePlayModuleTypes); j++ {
			idx := i + j
			moduleType := freePlayModuleTypes[idx]
			name := freePlayModuleNames[idx]
			enabled := m.freePlayConfig.EnabledModules[moduleType]
			checkbox := "[ ]"
			if enabled {
				checkbox = "[x]"
			}
			if m.freePlayInModules && m.freePlayCursor-4 == idx {
				row = append(row, styles.Active.Render("  "+checkbox+" "+name))
			} else {
				row = append(row, "  "+checkbox+" "+name)
			}
		}
		moduleCols = append(moduleCols, row)
	}

	for _, row := range moduleCols {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, row...))
	}

	rows = append(rows, "")
	rows = append(rows, "                     [ START GAME ]")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		rows...,
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		styles.HeaderBox.Render(styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES")),
		styles.ContentBox.Render(content),
		m.renderFooter(),
	)
}

func (m *Model) handleFreePlayAdvancedKeys(key string) (tea.Cmd, bool) {
	handled := true
	switch key {
	case "up", "k":
		if m.freePlayInModules {
			if m.freePlayCursor > 4 {
				m.freePlayCursor--
			} else {
				m.freePlayInModules = false
			}
		} else if m.freePlayCursor > 0 {
			m.freePlayCursor--
		}
	case "down", "j":
		if !m.freePlayInModules {
			if m.freePlayCursor < 3 {
				m.freePlayCursor++
			} else {
				m.freePlayInModules = true
				m.freePlayCursor = 4
			}
		} else if m.freePlayCursor < 4+len(freePlayModuleTypes) {
			m.freePlayCursor++
		}
	case "left", "h":
		if !m.freePlayInModules {
			switch m.freePlayCursor {
			case 0:
				if m.freePlayConfig.TimerSeconds > 30 {
					m.freePlayConfig.TimerSeconds -= 30
				}
			case 1:
				if m.freePlayConfig.MaxStrikes > 1 {
					m.freePlayConfig.MaxStrikes--
				}
			case 2:
				if m.freePlayConfig.NumFaces > 1 {
					m.freePlayConfig.NumFaces--
				}
			case 3:
				if m.freePlayConfig.ModulesPerFace > 1 {
					m.freePlayConfig.ModulesPerFace--
				}
			}
		}
	case "right", "l":
		if !m.freePlayInModules {
			switch m.freePlayCursor {
			case 0:
				if m.freePlayConfig.TimerSeconds < 3600 {
					m.freePlayConfig.TimerSeconds += 30
				}
			case 1:
				if m.freePlayConfig.MaxStrikes < 10 {
					m.freePlayConfig.MaxStrikes++
				}
			case 2:
				if m.freePlayConfig.NumFaces < 6 {
					m.freePlayConfig.NumFaces++
				}
			case 3:
				if m.freePlayConfig.ModulesPerFace < 12 {
					m.freePlayConfig.ModulesPerFace++
				}
			}
		}
	case " ":
		if m.freePlayInModules {
			idx := m.freePlayCursor - 4
			if idx >= 0 && idx < len(freePlayModuleTypes) {
				moduleType := freePlayModuleTypes[idx]
				m.freePlayConfig.EnabledModules[moduleType] = !m.freePlayConfig.EnabledModules[moduleType]
			}
		}
	case "enter":
		if m.freePlayInModules && m.freePlayCursor == 4+len(freePlayModuleTypes) {
			m.buildAndStartCustomGame()
		}
	case "esc":
		m.state = StateFreePlayMenu
		m.freePlaySelection = 0
	default:
		handled = false
	}
	return nil, handled
}

func (m *Model) buildAndStartCustomGame() {
	var modules []*pb.ModuleSpec
	for i, moduleType := range freePlayModuleTypes {
		if m.freePlayConfig.EnabledModules[moduleType] {
			modules = append(modules, &pb.ModuleSpec{
				Type:  freePlayModuleTypes[i],
				Count: int32(m.freePlayConfig.ModulesPerFace),
			})
		}
	}

	m.pendingGameConfig = &pb.GameConfig{
		ConfigType: &pb.GameConfig_Custom{
			Custom: &pb.CustomBombConfig{
				TimerSeconds:      int32(m.freePlayConfig.TimerSeconds),
				MaxStrikes:        int32(m.freePlayConfig.MaxStrikes),
				NumFaces:          int32(m.freePlayConfig.NumFaces),
				Modules:           modules,
				MaxModulesPerFace: int32(m.freePlayConfig.ModulesPerFace),
			},
		},
	}
	m.state = StateLoading
}
