// The bot went to another dimension: forget the blocks and the entities, write
// where it is, and answer as a client does: loading screen start, then a
// moment later loading screen end and "dimension change done". Without them
// the server keeps loading.
package arrive

import (
	"time"

	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// loadingTime was enough in the version 2 experiments; a client waits for the chunks.
const loadingTime = 5 * time.Second

func (a *Arrive) changeDimension(change *packet.ChangeDimension) {
	a.blocks.Forget()
	a.entities.Forget()
	a.self.SetDimension(change.Dimension)
	a.self.SetPosition(change.Position)
	// Nothing rides along to another dimension.
	a.self.SetVehicle(self.Vehicle{}, false)
	start := &packet.ServerBoundLoadingScreen{Type: packet.LoadingScreenTypeStart}
	end := &packet.ServerBoundLoadingScreen{Type: packet.LoadingScreenTypeEnd}
	if id, ok := change.LoadingScreenID.Value(); ok {
		start.LoadingScreenID = protocol.Option(id)
		end.LoadingScreenID = protocol.Option(id)
	}
	_ = a.link.SendServerBoundLoadingScreen(start)
	time.AfterFunc(loadingTime, func() {
		_ = a.link.SendServerBoundLoadingScreen(end)
		_ = a.link.SendPlayerAction(&packet.PlayerAction{EntityRuntimeID: a.self.EntityRuntimeID(), ActionType: protocol.PlayerActionDimensionChangeDone})
	})
}
