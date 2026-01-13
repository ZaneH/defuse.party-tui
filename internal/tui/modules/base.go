package modules

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ZaneH/keep-talking-tui/internal/client"
	pb "github.com/ZaneH/keep-talking-tui/proto"
)

type ModuleModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string

	ID() string
	ModuleType() pb.Module_ModuleType
	IsSolved() bool
	UpdateState(mod *pb.Module)

	Footer() string
}

func NewModule(mod *pb.Module, client client.GameClient, sessionID, bombID string) ModuleModel {
	switch mod.GetType() {
	case pb.Module_WIRES:
		return NewWiresModule(mod, client, sessionID, bombID)
	case pb.Module_BIG_BUTTON:
		return NewBigButtonModule(mod, client, sessionID, bombID)
	case pb.Module_KEYPAD:
		return NewKeypadModule(mod, client, sessionID, bombID)
	default:
		return NewUnimplementedModule(mod)
	}
}
