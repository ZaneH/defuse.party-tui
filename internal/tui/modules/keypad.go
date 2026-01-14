package modules

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	"github.com/ZaneH/keep-talking-tui/internal/styles"
	pb "github.com/ZaneH/keep-talking/pkg/proto"
)

type KeypadModule struct {
	mod       *pb.Module
	client    client.GameClient
	sessionID string
	bombID    string

	width  int
	height int

	activatedSymbols map[int]bool

	message     string
	messageType string
}

var symbolMap = map[pb.Symbol]string{
	pb.Symbol_COPYRIGHT:    "©",
	pb.Symbol_FILLEDSTAR:   "★",
	pb.Symbol_HOLLOWSTAR:   "☆",
	pb.Symbol_SMILEYFACE:   "☺",
	pb.Symbol_DOUBLEK:      "Ж",
	pb.Symbol_OMEGA:        "Ω",
	pb.Symbol_SQUIDKNIFE:   "Ѯ",
	pb.Symbol_PUMPKIN:      "Ѫ",
	pb.Symbol_HOOKN:        "Ҩ",
	pb.Symbol_SIX:          "б",
	pb.Symbol_SQUIGGLYN:    "Ҋ",
	pb.Symbol_AT:           "Ѧ",
	pb.Symbol_AE:           "Æ",
	pb.Symbol_MELTEDTHREE:  "Ӭ",
	pb.Symbol_EURO:         "€",
	pb.Symbol_NWITHHAT:     "Ñ",
	pb.Symbol_DRAGON:       "Ψ",
	pb.Symbol_QUESTIONMARK: "¿",
	pb.Symbol_PARAGRAPH:    "¶",
	pb.Symbol_RIGHTC:       "Ͽ",
	pb.Symbol_LEFTC:        "Ͼ",
	pb.Symbol_PITCHFORK:    "Ѱ",
	pb.Symbol_CURSIVE:      "ϗ",
	pb.Symbol_TRACKS:       "☰",
	pb.Symbol_BALLOON:      "Ѳ",
	pb.Symbol_UPSIDEDOWNY:  "λ",
	pb.Symbol_BT:           "Ƀ",
}

func NewKeypadModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) *KeypadModule {
	return &KeypadModule{
		mod:              mod,
		client:           client,
		sessionID:        sessionID,
		bombID:           bombID,
		activatedSymbols: make(map[int]bool),
	}
}

func (m *KeypadModule) Init() tea.Cmd {
	return nil
}

func (m *KeypadModule) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4":
			pos := int(msg.String()[0] - '1')
			return m, m.activateSymbol(pos)
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

func (m *KeypadModule) activateSymbol(pos int) tea.Cmd {
	return func() tea.Msg {
		state := m.mod.GetKeypadState()
		if state == nil {
			return ModuleResultMsg{Err: fmt.Errorf("no keypad state")}
		}

		symbols := state.GetDisplayedSymbols()
		if pos < 0 || pos >= len(symbols) {
			return ModuleResultMsg{Err: fmt.Errorf("invalid position")}
		}

		if m.activatedSymbols[pos] {
			return ModuleResultMsg{Err: fmt.Errorf("symbol already activated")}
		}

		symbol := symbols[pos]

		input := &pb.PlayerInput{
			SessionId: m.sessionID,
			BombId:    m.bombID,
			ModuleId:  m.mod.GetId(),
			Input:     &pb.PlayerInput_KeypadInput{KeypadInput: &pb.KeypadInput{Symbol: symbol}},
		}

		result, err := m.client.SendInput(context.Background(), input)
		if err != nil {
			return ModuleResultMsg{Err: err}
		}

		m.activatedSymbols[pos] = true

		if result.GetStrike() {
			m.message = "STRIKE! Wrong symbol!"
			m.messageType = "error"
			m.activatedSymbols = make(map[int]bool)
		} else if result.GetSolved() {
			m.message = "Module solved!"
			m.messageType = "success"
		} else {
			m.message = ""
			m.messageType = ""
		}

		return ModuleResultMsg{Result: result}
	}
}

func (m *KeypadModule) View() string {
	state := m.mod.GetKeypadState()
	if state == nil {
		return styles.Error.Render("No keypad state available")
	}

	symbols := state.GetDisplayedSymbols()
	if len(symbols) == 0 {
		return styles.Subtitle.Render("No symbols displayed")
	}

	var symbolLines []string
	for i := 0; i < len(symbols); i += 2 {
		leftSymbol := symbols[i]
		rightSymbol := symbols[i+1]

		leftChar := getSymbolChar(leftSymbol)
		rightChar := getSymbolChar(rightSymbol)

		leftActivated := m.activatedSymbols[i]
		rightActivated := m.activatedSymbols[i+1]

		leftBox := renderKeypadButton(leftChar, leftActivated, i+1)
		rightBox := renderKeypadButton(rightChar, rightActivated, i+2)

		row := lipgloss.JoinHorizontal(lipgloss.Center, leftBox, rightBox)
		symbolLines = append(symbolLines, row)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render("KEYPAD"),
		"",
		lipgloss.JoinVertical(lipgloss.Center, symbolLines...),
		"",
	)

	if m.message != "" {
		if m.messageType == "error" {
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				styles.Error.Render(m.message),
			)
		} else if m.messageType == "success" {
			content = lipgloss.JoinVertical(
				lipgloss.Center,
				content,
				styles.Success.Render(m.message),
			)
		}
	}

	return lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(content)
}

func renderKeypadButton(symbol string, activated bool, position int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(8).
		Height(3)

	if activated {
		boxStyle = boxStyle.BorderForeground(lipgloss.Color("#6BCB77"))
	}

	symbolStyle := lipgloss.NewStyle().
		Width(8).
		Height(3).
		Align(lipgloss.Center).
		Bold(true)

	if activated {
		symbolStyle = symbolStyle.Foreground(lipgloss.Color("#6BCB77"))
	}

	positionLabel := fmt.Sprintf("[%d]", position)

	boxContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Help.Render(positionLabel),
		symbolStyle.Render(symbol),
	)

	if activated {
		boxContent = lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Success.Render(positionLabel),
			symbolStyle.Render(symbol),
			styles.Success.Render("✓"),
		)
	}

	return boxStyle.Render(boxContent)
}

func getSymbolChar(s pb.Symbol) string {
	if char, ok := symbolMap[s]; ok {
		return char
	}
	return "?"
}

func (m *KeypadModule) ID() string {
	return m.mod.GetId()
}

func (m *KeypadModule) ModuleType() pb.Module_ModuleType {
	return pb.Module_KEYPAD
}

func (m *KeypadModule) IsSolved() bool {
	return m.mod.GetSolved()
}

func (m *KeypadModule) UpdateState(mod *pb.Module) {
	if mod.GetSolved() {
		m.mod.Solved = true
	}
}

func (m *KeypadModule) Footer() string {
	return "[1-4] Select symbol | [ESC] Back to bomb"
}
