// One block changed together with a falling block entity: the block part is
// the same as update_block.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) updateBlockSynced(update *packet.UpdateBlockSynced) {
	a.blocks.SetBlock(update.Position, int(update.Layer), a.block(update.NewBlockRuntimeID))
}
