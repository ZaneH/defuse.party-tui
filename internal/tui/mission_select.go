package tui

import (
	"github.com/ZaneH/defuse.party-tui/internal/styles"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MissionSection struct {
	Name     string
	Missions []MissionInfo
}

type MissionInfo struct {
	Name    string
	Mission pb.Mission
}

var missionSections = []MissionSection{
	{
		Name: "Section 1: Introduction",
		Missions: []MissionInfo{
			{Name: "The First Bomb", Mission: pb.Mission_THE_FIRST_BOMB},
		},
	},
	{
		Name: "Section 2: The Basics",
		Missions: []MissionInfo{
			{Name: "Something Old, Something New", Mission: pb.Mission_SOMETHING_OLD_SOMETHING_NEW},
			{Name: "Double Your Money", Mission: pb.Mission_DOUBLE_YOUR_MONEY},
			{Name: "One Step Up", Mission: pb.Mission_ONE_STEP_UP},
			{Name: "Pick Up The Pace", Mission: pb.Mission_PICK_UP_THE_PACE},
		},
	},
	{
		Name: "Section 3: Moderate",
		Missions: []MissionInfo{
			{Name: "A Hidden Message", Mission: pb.Mission_A_HIDDEN_MESSAGE},
			{Name: "Something's Different", Mission: pb.Mission_SOMETHINGS_DIFFERENT},
			{Name: "One Giant Leap", Mission: pb.Mission_ONE_GIANT_LEAP},
			{Name: "Fair Game", Mission: pb.Mission_FAIR_GAME},
			{Name: "Pick Up The Pace II", Mission: pb.Mission_PICK_UP_THE_PACE_II},
			{Name: "No Room For Error", Mission: pb.Mission_NO_ROOM_FOR_ERROR},
			{Name: "Eight Minutes", Mission: pb.Mission_EIGHT_MINUTES},
		},
	},
	{
		Name: "Section 4: Needy Modules",
		Missions: []MissionInfo{
			{Name: "A Small Wrinkle", Mission: pb.Mission_A_SMALL_WRINKLE},
			{Name: "Pay Attention", Mission: pb.Mission_PAY_ATTENTION},
			{Name: "The Knob", Mission: pb.Mission_THE_KNOB},
			{Name: "Multitasker", Mission: pb.Mission_MULTI_TASKER},
		},
	},
	{
		Name: "Section 5: Challenging",
		Missions: []MissionInfo{
			{Name: "Wires Wires Everywhere", Mission: pb.Mission_WIRES_WIRES_EVERYWHERE},
			{Name: "Computer Hacking", Mission: pb.Mission_COMPUTER_HACKING},
			{Name: "Who's On First Challenge", Mission: pb.Mission_WHOS_ON_FIRST_CHALLENGE},
			{Name: "Fiendish", Mission: pb.Mission_FIENDISH},
			{Name: "Pick Up The Pace III", Mission: pb.Mission_PICK_UP_THE_PACE_III},
			{Name: "One With Everything", Mission: pb.Mission_ONE_WITH_EVERYTHING},
		},
	},
	{
		Name: "Section 6: Extreme",
		Missions: []MissionInfo{
			{Name: "Pick Up The Pace IV", Mission: pb.Mission_PICK_UP_THE_PACE_IV},
			{Name: "Juggler", Mission: pb.Mission_JUGGLER},
			{Name: "Double Trouble", Mission: pb.Mission_DOUBLE_TROUBLE},
			{Name: "I Am Hardcore", Mission: pb.Mission_I_AM_HARDCORE},
		},
	},
	{
		Name: "Section 7: Exotic",
		Missions: []MissionInfo{
			{Name: "Blinkenlights", Mission: pb.Mission_BLINKENLIGHTS},
			{Name: "Applied Theory", Mission: pb.Mission_APPLIED_THEORY},
			{Name: "A Maze Ing", Mission: pb.Mission_A_MAZE_ING},
			{Name: "Snip Snap", Mission: pb.Mission_SNIP_SNAP},
			{Name: "Rainbow Table", Mission: pb.Mission_RAINBOW_TABLE},
			{Name: "Blinkenlights II", Mission: pb.Mission_BLINKENLIGHTS_II},
		},
	},
}

func (m *Model) sectionSelectView() string {
	var items []string
	for i, section := range missionSections {
		if i == m.sectionSelection {
			items = append(items, styles.Active.Render("> "+section.Name))
		} else {
			items = append(items, "  "+section.Name)
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("SELECT SECTION"),
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

func (m *Model) missionSelectView() string {
	section := missionSections[m.sectionSelection]
	var items []string
	for i, mission := range section.Missions {
		if i == m.missionSelection {
			items = append(items, styles.Active.Render("> "+mission.Name))
		} else {
			items = append(items, "  "+mission.Name)
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render(section.Name),
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

func (m *Model) handleSectionSelectKeys(key string) (tea.Cmd, bool) {
	handled := true
	switch key {
	case "up", "k":
		if m.sectionSelection > 0 {
			m.sectionSelection--
		}
	case "down", "j":
		if m.sectionSelection < len(missionSections)-1 {
			m.sectionSelection++
		}
	case "enter":
		m.state = StateMissionSelect
		m.missionSelection = 0
	case "esc":
		m.state = StateMainMenu
		m.menuSelection = 0
	default:
		handled = false
	}
	return nil, handled
}

func (m *Model) handleMissionSelectKeys(key string) (tea.Cmd, bool) {
	section := missionSections[m.sectionSelection]
	handled := true
	switch key {
	case "up", "k":
		if m.missionSelection > 0 {
			m.missionSelection--
		}
	case "down", "j":
		if m.missionSelection < len(section.Missions)-1 {
			m.missionSelection++
		}
	case "enter":
		mission := section.Missions[m.missionSelection]
		m.pendingGameConfig = &pb.GameConfig{
			ConfigType: &pb.GameConfig_Preset{
				Preset: &pb.PresetMissionConfig{
					Mission: mission.Mission,
				},
			},
		}
		m.state = StateLoading
		return m.StartGame(m.pendingGameConfig), true
	case "esc":
		m.state = StateSectionSelect
		m.missionSelection = 0
	default:
		handled = false
	}
	return nil, handled
}
