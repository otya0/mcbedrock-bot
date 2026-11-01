// The server set the bot's health: write it, keeping the most it can be.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) setHealth(changed *packet.SetHealth) {
	_, maximum := a.self.Health()
	a.self.SetHealth(float32(changed.Health), maximum)
}
