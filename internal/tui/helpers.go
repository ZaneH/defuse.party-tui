package tui

import (
	"strings"
	"time"

	"github.com/ZaneH/defuse.party-tui/internal/styles"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
)

type FreePlayConfig struct {
	TimerSeconds   int
	MaxStrikes     int
	NumFaces       int
	ModulesPerFace int

	EnabledModules map[pb.Module_ModuleType]bool
}

func DefaultFreePlayConfig() FreePlayConfig {
	return FreePlayConfig{
		TimerSeconds:   300,
		MaxStrikes:     3,
		NumFaces:       2,
		ModulesPerFace: 6,
		EnabledModules: map[pb.Module_ModuleType]bool{
			pb.Module_WIRES:          true,
			pb.Module_PASSWORD:       true,
			pb.Module_BIG_BUTTON:     true,
			pb.Module_SIMON:          true,
			pb.Module_KEYPAD:         true,
			pb.Module_WHOS_ON_FIRST:  true,
			pb.Module_MEMORY:         true,
			pb.Module_MORSE:          true,
			pb.Module_MAZE:           true,
			pb.Module_NEEDY_VENT_GAS: false,
			pb.Module_NEEDY_KNOB:     false,
		},
	}
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

func formatTimer(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return strings.TrimPrefix(formatTwoDigits(minutes), "0") + ":" + formatTwoDigits(seconds)
}

func formatTwoDigits(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10%10)) + string(rune('0'+n%10))
}

func formatTimeString(seconds int) string {
	mins := seconds / 60
	secs := seconds % 60
	return formatTwoDigits(mins) + ":" + formatTwoDigits(secs)
}

func hyperlink(url, text string) string {
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}

func (m *Model) renderFooter() string {
	hint := ""
	switch m.state {
	case StateMainMenu:
		hint = "[↑/↓] Navigate  [ENTER] Select  [Q] Quit"
	case StateSectionSelect:
		hint = "[↑/↓] Navigate  [ENTER] Select section  [ESC] Back"
	case StateMissionSelect:
		hint = "[↑/↓] Navigate  [ENTER] Start mission  [ESC] Back to sections"
	case StateFreePlayMenu:
		hint = "[↑/↓] Navigate  [ENTER] Select  [ESC] Back"
	case StateFreePlayAdvanced:
		hint = "[↑/↓] Navigate  [←/→] Adjust  [SPACE] Toggle  [ENTER] Start  [ESC] Back"
	case StateBombSelection:
		hint = "[ENTER] Pick up bomb | [↑/↓] Navigate | [Q]uit"
	case StateBombView:
		hint = "[1-9] Select module | [<]/[>] Flip face | [ESC] Put down | [Q]uit"
	case StateModuleActive:
		hint = m.activeModule.Footer()
	case StateGameOver:
		hint = "[↑/↓] Navigate  [ENTER] Select"
	}
	return styles.FooterBox.Render(styles.Help.Render(hint))
}
