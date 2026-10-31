// A chunk is listed, with its biomes. In request mode (every chunk so far)
// its blocks come as sub_chunk once the bot asks, which it does for the chunks
// near it, as a client does by itself. Otherwise the payload carries the sub-chunks.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// askRadius is how far around the centre, in chunks, the bot asks for blocks.
const askRadius = 4

func (a *Arrive) levelChunk(chunk *packet.LevelChunk) {
	position := [2]int32{chunk.Position[0], chunk.Position[1]}
	lowest, highest := subChunkRange(chunk.Dimension)
	if limit, requestMode := chunk.SubChunkLimit.Value(); requestMode {
		// limit is the highest sub-chunk that is not all air, counted from the
		// lowest (2026-10-21 experiments); -1 is all. The payload is the biomes.
		top := int32(-1)
		if limit >= 0 {
			top = lowest + limit
		}
		a.blocks.SetListed(position, top, false)
		if biomes, err := readBiomes(chunk.RawPayload, int(highest-lowest+1)); err == nil {
			a.blocks.SetBiomes(position, lowest, a.namedBiomes(biomes))
		}
		a.askForNearChunks()
		return
	}
	a.blocks.SetListed(position, -1, true)
	if chunk.CacheEnabled {
		return
	}
	data := chunk.RawPayload
	for i := uint32(0); i < chunk.SubChunkCount; i++ {
		sub, yIndex, hasYIndex, used, err := readSubChunk(data)
		if err != nil {
			return
		}
		y := lowest + int32(i)
		if hasYIndex {
			y = int32(yIndex)
		}
		if y < lowest || y > highest {
			return
		}
		a.blocks.SetSubChunk([3]int32{position[0], y, position[1]}, a.namedSubChunk(sub))
		data = data[used:]
	}
}

// askForNearChunks sends one sub_chunk_request per listed chunk near the centre
// not asked for yet, naming its sub-chunks from the lowest to its highest one.
func (a *Arrive) askForNearChunks() {
	dimension := a.self.Dimension()
	lowest, highest := subChunkRange(dimension)
	centre, chunks, ok := a.blocks.TakeChunksToAsk(askRadius, lowest, highest)
	if !ok {
		return
	}
	for _, chunk := range chunks {
		top := highest
		if chunk.Highest >= 0 && chunk.Highest < top {
			top = chunk.Highest
		}
		var offsets []protocol.SubChunkOffset
		for y := lowest; y <= top; y++ {
			offsets = append(offsets, protocol.SubChunkOffset{int8(chunk.Position[0] - centre[0]), int8(y), int8(chunk.Position[1] - centre[1])})
		}
		request := &packet.SubChunkRequest{Dimension: dimension, Position: protocol.SubChunkPos{centre[0], 0, centre[1]}, Offsets: offsets}
		if a.link.SendSubChunkRequest(request) != nil {
			return
		}
	}
}

// subChunkRange is the lowest and highest sub-chunk of a dimension: the
// overworld y -64 to 319, the nether 0 to 127, the end 0 to 255.
func subChunkRange(dimension int32) (int32, int32) {
	switch dimension {
	case packet.DimensionNether:
		return 0, 7
	case packet.DimensionEnd:
		return 0, 15
	}
	return -4, 19
}

// namedBiomes gives each biome number of the layers its name.
func (a *Arrive) namedBiomes(layers []packedLayer) []blocks.BiomeLayer {
	named := make([]blocks.BiomeLayer, len(layers))
	for i, layer := range layers {
		named[i].Indices = layer.indices
		for _, id := range layer.palette {
			named[i].Palette = append(named[i].Palette, a.names.Biome(id))
		}
	}
	return named
}
