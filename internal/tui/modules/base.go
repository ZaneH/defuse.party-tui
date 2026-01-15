package modules

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ZaneH/defuse.party-tui/internal/client"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
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
	case pb.Module_CLOCK:
		return NewUnimplementedModule(mod)
	case pb.Module_WIRES:
		return NewWiresModule(mod, client, sessionID, bombID)
	case pb.Module_BIG_BUTTON:
		return NewBigButtonModule(mod, client, sessionID, bombID)
	case pb.Module_KEYPAD:
		return NewKeypadModule(mod, client, sessionID, bombID)
	case pb.Module_PASSWORD:
		return NewPasswordModule(mod, client, sessionID, bombID)
	case pb.Module_MORSE:
		return NewMorseModule(mod, client, sessionID, bombID)
	case pb.Module_SIMON:
		return NewSimonModule(mod, client, sessionID, bombID)
	case pb.Module_MEMORY:
		return NewMemoryModule(mod, client, sessionID, bombID)
	case pb.Module_WHOS_ON_FIRST:
		return NewWhosOnFirstModule(mod, client, sessionID, bombID)
	case pb.Module_MAZE:
		return NewMazeModule(mod, client, sessionID, bombID)
	case pb.Module_NEEDY_VENT_GAS:
		return NewNeedyVentGasModule(mod, client, sessionID, bombID)
	case pb.Module_NEEDY_KNOB:
		return NewNeedyKnobModule(mod, client, sessionID, bombID)
	default:
		return NewUnimplementedModule(mod)
	}
}
