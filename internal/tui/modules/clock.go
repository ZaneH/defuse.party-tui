package modules

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/defuse.party-tui/internal/styles"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
)

type ClockModule struct {
	mod        *pb.Module
	startedAt  time.Time
	duration   time.Duration
	strikes    int32
	maxStrikes int32
	width      int
	height     int
}

// 7-segment digit patterns - each digit is 6 lines tall
var segmentDigits = map[rune][]string{
	'0': {
		" ███╗ ",
		"████║ ",
		"╚═██║ ",
		"  ██║ ",
		"  ██║ ",
		"  ╚═╝ ",
	},
	'1': {
		" ██╗",
		" ██║",
		" ██║",
		" ██║",
		" ██║",
		" ╚═╝",
	},
	'2': {
		"██████╗ ",
		"╚════██╗",
		" █████╔╝",
		"██╔═══╝ ",
		"███████╗",
		"╚══════╝",
	},
	'3': {
		"██████╗ ",
		"╚════██╗",
		" █████╔╝",
		" ╚═══██╗",
		"██████╔╝",
		"╚═════╝ ",
	},
	'4': {
		"██╗  ██╗",
		"██║  ██║",
		"███████║",
		"╚════██║",
		"     ██║",
		"     ╚═╝",
	},
	'5': {
		"███████╗",
		"██╔════╝",
		"███████╗",
		"╚════██║",
		"███████║",
		"╚══════╝",
	},
	'6': {
		" ██████╗ ",
		"██╔════╝ ",
		"███████╗ ",
		"██╔═══██╗",
		"╚██████╔╝",
		" ╚═════╝ ",
	},
	'7': {
		"███████╗",
		"╚════██║",
		"    ██╔╝",
		"   ██╔╝ ",
		"   ██║  ",
		"   ╚═╝  ",
	},
	'8': {
		" █████╗ ",
		"██╔═══██╗",
		"╚█████╔╝",
		"██╔═══██╗",
		"╚█████╔╝",
		" ╚════╝ ",
	},
	'9': {
		" ██████╗ ",
		"██╔═══██╗",
		"╚██████║ ",
		" ╚═══██║ ",
		"██████╔╝ ",
		"╚═════╝  ",
	},
	':': {
		"   ",
		" ╔╗",
		" ╚╝",
		" ╔╗",
		" ╚╝",
		"   ",
	},
}

func NewClockModule(mod *pb.Module, startedAt time.Time, duration time.Duration, strikes, maxStrikes int32) *ClockModule {
	return &ClockModule{
		mod:        mod,
		startedAt:  startedAt,
		duration:   duration,
		strikes:    strikes,
		maxStrikes: maxStrikes,
	}
}

func (m *ClockModule) Init() tea.Cmd {
	return nil
}

func (m *ClockModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
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

func (m *ClockModule) View() string {
	elapsed := time.Since(m.startedAt)
	remaining := m.duration - elapsed
	if remaining < 0 {
		remaining = 0
	}

	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60

	timerStyle := styles.Normal
	if remaining < 30*time.Second {
		timerStyle = styles.Error
	} else if remaining < 60*time.Second {
		timerStyle = styles.Warning
	}

	timerText := fmt.Sprintf("%02d:%02d", minutes, seconds)
	largeTimer := m.renderLargeTimer(timerText)

	coloredTimer := timerStyle.Render(largeTimer)

	timerDisplay := lipgloss.NewStyle().
		Padding(2, 0).
		Align(lipgloss.Center).
		Render(coloredTimer)
	strikesDisplay := ""
	for i := int32(0); i < m.maxStrikes; i++ {
		if i < m.strikes {
			strikesDisplay += "[X] "
		} else {
			strikesDisplay += "[ ] "
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("CLOCK"),
		"",
		timerDisplay,
		"",
		styles.Normal.Render("STRIKES: "+strikesDisplay),
		"",
		styles.Help.Render("This is a display-only module."),
	)

	return lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(content)
}

func (m *ClockModule) renderLargeTimer(text string) string {
	// Get digit patterns for each character
	var digitLines [][]string
	for _, char := range text {
		if pattern, exists := segmentDigits[char]; exists {
			digitLines = append(digitLines, pattern)
		}
	}

	// Combine horizontally
	var lines []string
	for lineIdx := 0; lineIdx < 6; lineIdx++ {
		var lineBuilder strings.Builder
		for _, digit := range digitLines {
			lineBuilder.WriteString(digit[lineIdx])
			lineBuilder.WriteString(" ")
		}
		lines = append(lines, lineBuilder.String())
	}

	return strings.Join(lines, "\n")
}

func (m *ClockModule) ID() string {
	return m.mod.GetId()
}

func (m *ClockModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_CLOCK
}

func (m *ClockModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *ClockModule) UpdateState(mod *pb.Module) {
	m.mod = mod
}

func (m *ClockModule) UpdateStrikes(strikes int32) {
	m.strikes = strikes
}

func (m *ClockModule) Footer() string {
	return "[ESC] Back to bomb"
}
