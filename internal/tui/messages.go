package tui

import (
	"time"

	"github.com/ZaneH/defuse.party-tui/internal/client"
	pb "github.com/ZaneH/defuse.party-go/pkg/proto"
)

type loadingErrorMsg struct{ err error }

type gameReadyMsg struct {
	client    client.GameClient
	sessionID string
	bombs     []*pb.Bomb
}

type tickMsg struct{ t time.Time }

type startGameMsg struct {
	config *pb.GameConfig
}

type returnToMenuMsg struct{}

type showManualMsg struct{}
