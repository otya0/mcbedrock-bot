// Package blocks keeps the blocks and biomes around the bot by position, as
// the server packs them: for each 16x16x16 sub-chunk, a palette and an index
// per block. Each palette entry carries the block's name, so At answers what
// is at a position.
package blocks

import (
	"maps"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AirRuntimeID is minecraft:air as a hashed block network ID, the block of a
// sub-chunk the server sent as all air.
const AirRuntimeID = 3690217760

// Air is the block of a sub-chunk the server sent as all air, and of a new
// layer 1 (the water inside blocks) before anything is set in it.
func Air() Block {
	return Block{RuntimeID: AirRuntimeID, Name: "minecraft:air"}
}

// Block is what is at a position: the number the server sends, the name and
// property values from Mojang's data ("" when not listed there), and the
// biome ("" until the server has sent the chunk's biomes).
type Block struct {
	RuntimeID uint32
	Name      string
	States    map[string]string
	Biome     string
}

// Layer is one storage layer of a sub-chunk: 0 the blocks, 1 the water inside
// them. Block (x, y, z) of the sub-chunk is Palette[Indices[x*256 + z*16 + y]];
// Indices is nil when the whole layer is Palette[0].
type Layer struct {
	Palette []Block
	Indices []uint16
}

type SubChunk struct {
	Layers []Layer
}

// BiomeLayer is the biomes of one sub-chunk, by name, packed as a Layer is.
type BiomeLayer struct {
	Palette []string
	Indices []uint16
}

// Chunk is a chunk the server listed, to be asked for its sub-chunks.
type Chunk struct {
	Position [2]int32
	Highest  int32
}

type listedChunk struct {
	highest int32
	asked   bool
	// The biomes come with the listing, from the dimension's lowest sub-chunk up.
	lowest int32
	biomes []BiomeLayer
}

type Blocks struct {
	mu          sync.Mutex
	subChunks   map[[3]int32]*SubChunk
	listed      map[[2]int32]*listedChunk
	centre      [2]int32
	loadRadius  int32
	centreKnown bool
}

func New() *Blocks {
	return &Blocks{subChunks: map[[3]int32]*SubChunk{}, listed: map[[2]int32]*listedChunk{}}
}

// At gives the block at a position (storage layer 0) with its biome, if the
// bot has its sub-chunk.
func (b *Blocks) At(position protocol.BlockPos) (Block, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	sub, ok := b.subChunks[subChunkOf(position)]
	if !ok || len(sub.Layers) == 0 {
		return Block{}, false
	}
	found := sub.Layers[0].at(indexWithin(position))
	found.States = maps.Clone(found.States)
	found.Biome = b.biomeAtLocked(position)
	return found, true
}

func (l *Layer) at(index int) Block {
	if l.Indices == nil {
		return l.Palette[0]
	}
	return l.Palette[l.Indices[index]]
}

func (b *Blocks) biomeAtLocked(position protocol.BlockPos) string {
	chunk, ok := b.listed[[2]int32{position[0] >> 4, position[2] >> 4}]
	if !ok {
		return ""
	}
	i := int(position[1]>>4 - chunk.lowest)
	if i < 0 || i >= len(chunk.biomes) {
		return ""
	}
	layer := chunk.biomes[i]
	if layer.Indices == nil {
		return layer.Palette[0]
	}
	return layer.Palette[layer.Indices[indexWithin(position)]]
}

// SetSubChunk keeps a sub-chunk of a listed chunk; one of a chunk the server
// never listed is dropped.
func (b *Blocks) SetSubChunk(position [3]int32, sub SubChunk) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, listed := b.listed[[2]int32{position[0], position[2]}]; !listed {
		return
	}
	b.subChunks[position] = &sub
}

// SetBlock changes one block. A block of a sub-chunk the bot does not have is
// left: the sub-chunk comes with it when the bot asks.
func (b *Blocks) SetBlock(position protocol.BlockPos, layer int, block Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	sub, ok := b.subChunks[subChunkOf(position)]
	if !ok || layer < 0 || layer > 1 {
		return
	}
	for len(sub.Layers) <= layer {
		sub.Layers = append(sub.Layers, Layer{Palette: []Block{Air()}})
	}
	changed := &sub.Layers[layer]
	index := -1
	for i, entry := range changed.Palette {
		if entry.RuntimeID == block.RuntimeID {
			index = i
			break
		}
	}
	if index == 0 && changed.Indices == nil {
		return
	}
	if index < 0 {
		changed.Palette = append(changed.Palette, block)
		index = len(changed.Palette) - 1
	}
	if changed.Indices == nil {
		changed.Indices = make([]uint16, 4096)
	}
	changed.Indices[indexWithin(position)] = uint16(index)
}

// SetListed remembers a chunk of level_chunk: highest is its highest sub-chunk
// that is not all air (-1: not said), asked whether it needs no asking.
func (b *Blocks) SetListed(chunk [2]int32, highest int32, asked bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listed[chunk] = &listedChunk{highest: highest, asked: asked}
}

// SetBiomes keeps the biomes of a listed chunk, from its lowest sub-chunk up.
func (b *Blocks) SetBiomes(chunk [2]int32, lowest int32, biomes []BiomeLayer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if listed, ok := b.listed[chunk]; ok {
		listed.lowest, listed.biomes = lowest, biomes
	}
}

// SetCentre is the chunk the server loads chunks around, and how far, in
// chunks: beyond that the server forgets what it sent and lists it again.
func (b *Blocks) SetCentre(chunk [2]int32, loadRadius int32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.centre, b.loadRadius, b.centreKnown = chunk, loadRadius, true
}

// TakeChunksToAsk gives the listed chunks within radius of the centre not yet
// asked for, and marks them asked. Chunks farther than radius+1 lose their
// sub-chunks (lowest to highest) so that they are asked for again when near;
// the listing stays, because the server lists a chunk only once. Chunks
// beyond the server's own radius are forgotten whole: it lists them again.
func (b *Blocks) TakeChunksToAsk(radius, lowest, highest int32) ([2]int32, []Chunk, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.centreKnown {
		return b.centre, nil, false
	}
	var chunks []Chunk
	for position, chunk := range b.listed {
		far := max(absolute(position[0]-b.centre[0]), absolute(position[1]-b.centre[1]))
		switch {
		case far <= radius && !chunk.asked:
			chunk.asked = true
			chunks = append(chunks, Chunk{Position: position, Highest: chunk.highest})
		case far > radius+1 && chunk.asked:
			for y := lowest; y <= highest; y++ {
				delete(b.subChunks, [3]int32{position[0], y, position[1]})
			}
			chunk.asked = false
		}
		if far > b.loadRadius+1 {
			delete(b.listed, position)
		}
	}
	return b.centre, chunks, true
}

// Forget drops every chunk: the dimension changed.
func (b *Blocks) Forget() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subChunks = map[[3]int32]*SubChunk{}
	b.listed = map[[2]int32]*listedChunk{}
	b.centreKnown = false
}

func subChunkOf(position protocol.BlockPos) [3]int32 {
	return [3]int32{position[0] >> 4, position[1] >> 4, position[2] >> 4}
}

func indexWithin(position protocol.BlockPos) int {
	return int(position[0]&15)<<8 | int(position[2]&15)<<4 | int(position[1]&15)
}

func absolute(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
