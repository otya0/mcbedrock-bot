// Several blocks of one sub-chunk changed at once: Blocks are layer 0, Extra 1.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) updateSubChunkBlocks(update *packet.UpdateSubChunkBlocks) {
	for _, entry := range update.Blocks {
		a.blocks.SetBlock(entry.BlockPos, 0, a.block(entry.BlockRuntimeID))
	}
	for _, entry := range update.Extra {
		a.blocks.SetBlock(entry.BlockPos, 1, a.block(entry.BlockRuntimeID))
	}
}
