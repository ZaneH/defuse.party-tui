package tui

import (
	"fmt"
	"time"

	"github.com/ZaneH/defuse.party-tui/internal/styles"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
	"github.com/charmbracelet/lipgloss"
)

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

		moduleLine := fmt.Sprintf("[%d] %s", i+1, modTypeName)
		if timer != "" {
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

func (m *Model) gameOverView() string {
	var title string
	if m.err != nil {
		if m.err.Error() == "BOOM! The bomb exploded." {
			title = "GAME OVER"
		} else if m.err.Error() == "time's up!" {
			title = "TIME'S UP!"
		} else {
			title = "GAME OVER"
		}
	} else {
		title = "CONGRATULATIONS!"
	}

	errMsg := ""
	if m.err != nil {
		errMsg = m.err.Error()
	}

	options := []string{"RETURN TO MENU", "QUIT"}
	var optionLines []string
	for i, opt := range options {
		if i == m.gameOverSelection {
			optionLines = append(optionLines, styles.Active.Render("> "+opt))
		} else {
			optionLines = append(optionLines, "  "+opt)
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render(title),
		"",
		styles.Subtitle.Render(errMsg),
		"",
		"",
		lipgloss.JoinVertical(lipgloss.Center, optionLines...),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		styles.HeaderBox.Render(styles.Title.Render("KEEP TALKING AND NOBODY EXPLODES")),
		styles.ContentBox.Render(content),
		m.renderFooter(),
	)
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
				return ""
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
				return ""
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
