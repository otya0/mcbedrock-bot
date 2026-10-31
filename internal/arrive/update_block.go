// One block changed (placed, broken, a door opened). Layer 1 is water in it.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) updateBlock(update *packet.UpdateBlock) {
	a.blocks.SetBlock(update.Position, int(update.Layer), a.block(update.NewBlockRuntimeID))
}
