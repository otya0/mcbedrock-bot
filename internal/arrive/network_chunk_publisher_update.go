// The server says which point it loads chunks around: ask for the chunks that
// came near.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) networkChunkPublisherUpdate(update *packet.NetworkChunkPublisherUpdate) {
	// The radius comes in blocks; 16 blocks make a chunk.
	radiusInChunks := int32((update.Radius + 15) / 16)
	a.blocks.SetCentre([2]int32{update.Position[0] >> 4, update.Position[2] >> 4}, radiusInChunks)
	a.askForNearChunks()
}
