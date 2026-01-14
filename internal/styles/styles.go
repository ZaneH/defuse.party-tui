package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Title      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B6B")).Padding(0, 1)
	Subtitle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	Help       = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	Warning    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	Error      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	Success    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6BCB77"))
	ContentBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(70)
	HeaderBox  = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).Padding(0, 1).Width(70)
	FooterBox  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Width(70)
	Normal     = lipgloss.NewStyle()
	Solved     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6BCB77")).Bold(true)
	Pending    = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	Active     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true)
	Strike     = lipgloss.NewStyle().Background(lipgloss.Color("#FF4444")).Foreground(lipgloss.Color("#FFFFFF"))
)

var (
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	Blue   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4DABF7"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	Black  = lipgloss.NewStyle().Foreground(lipgloss.Color("#212529"))
	White  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6BCB77"))
)

func Center(s string, width, height int) string {
	vWidth := lipgloss.Width(s)
	vHeight := lipgloss.Height(s)

	if vWidth >= width {
		return s
	}

	horizontal := (width - vWidth) / 2
	if horizontal < 0 {
		horizontal = 0
	}

	if vHeight >= height {
		return lipgloss.NewStyle().Width(width).Render(s)
	}

	vertical := (height - vHeight) / 2
	if vertical < 0 {
		vertical = 0
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(vertical, 0).
		Render(lipgloss.NewStyle().PaddingLeft(horizontal).Render(s))
}
