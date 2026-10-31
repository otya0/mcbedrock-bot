// The bot died: it stays dead until the respawn door presses the button and
// the server sends where to come back. The left button is let go of. The
// game's server says "may come back" after this packet, never before.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) deathInfo(*packet.DeathInfo) {
	a.self.SetDead(true)
	a.self.SetCanRespawn(false)
	a.self.SetBreaking(self.Breaking{}, false)
}
