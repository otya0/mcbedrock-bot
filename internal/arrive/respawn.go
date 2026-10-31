// While dead, the server says the bot may come back (state SearchingForSpawn),
// and after the button is pressed, where to (state ReadyToSpawn): write them,
// and tell the server the bot is back, as a client does by itself.
package arrive

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) respawn(respawn *packet.Respawn) {
	if !a.self.Dead() {
		return
	}
	switch respawn.State {
	case packet.RespawnStateSearchingForSpawn:
		a.self.SetCanRespawn(true)
	case packet.RespawnStateReadyToSpawn:
		a.self.SetDead(false)
		a.self.SetCanRespawn(false)
		a.self.SetPosition(respawn.Position)
		_ = a.link.SendPlayerAction(&packet.PlayerAction{EntityRuntimeID: a.self.EntityRuntimeID(), ActionType: protocol.PlayerActionRespawn})
	}
}
