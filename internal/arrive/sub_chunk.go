// The sub-chunks the bot asked for arrived: keep them, each block with its
// name. This file also reads the network form of a sub-chunk and of the
// biomes, which level_chunk uses too.
package arrive

import (
	"errors"
	"fmt"

	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// maxStorageLayers is how many storage layers a sub-chunk has in the game: the
// blocks, and the water inside them.
const maxStorageLayers = 2

func (a *Arrive) subChunk(answer *packet.SubChunk) {
	lowest, highest := subChunkRange(answer.Dimension)
	for _, entry := range answer.SubChunkEntries {
		position := [3]int32{answer.Position[0] + int32(entry.Offset[0]), answer.Position[1] + int32(entry.Offset[1]), answer.Position[2] + int32(entry.Offset[2])}
		if position[1] < lowest || position[1] > highest {
			continue
		}
		switch entry.Result {
		case protocol.SubChunkResultSuccessAllAir:
			a.blocks.SetSubChunk(position, blocks.SubChunk{Layers: []blocks.Layer{{Palette: []blocks.Block{blocks.Air()}}}})
		case protocol.SubChunkResultSuccess:
			data, _ := entry.RawPayload.Value()
			layers, yIndex, hasYIndex, _, err := readSubChunk(data)
			if err != nil {
				continue
			}
			if hasYIndex {
				position[1] = int32(yIndex)
			}
			if position[1] < lowest || position[1] > highest {
				continue
			}
			a.blocks.SetSubChunk(position, a.namedSubChunk(layers))
		}
	}
}

// packedLayer is one storage layer as it arrives: block runtime IDs (or biome
// numbers) and, when there is more than one, which entry each of the 4096 is.
type packedLayer struct {
	palette []uint32
	indices []uint16
}

// namedSubChunk gives each block of the palettes its name.
func (a *Arrive) namedSubChunk(layers []packedLayer) blocks.SubChunk {
	var sub blocks.SubChunk
	for _, layer := range layers {
		named := blocks.Layer{Indices: layer.indices}
		for _, runtimeID := range layer.palette {
			named.Palette = append(named.Palette, a.block(runtimeID))
		}
		sub.Layers = append(sub.Layers, named)
	}
	return sub
}

// block is a block runtime ID with its name and property values.
func (a *Arrive) block(runtimeID uint32) blocks.Block {
	name, states := a.names.Block(runtimeID)
	return blocks.Block{RuntimeID: runtimeID, Name: name, States: states}
}

// The network form, as read in the version 2 experiments and checked against
// 54 broken blocks: version (1, 8 or 9), a layer count (8, 9), a y index (9),
// then each layer: header (bits per block << 1 | 1), packed indices, palette.
func readSubChunk(data []byte) (layers []packedLayer, yIndex int8, hasYIndex bool, used int, err error) {
	reader := &byteReader{data: data}
	version, err := reader.byte()
	if err != nil {
		return nil, 0, false, 0, err
	}
	count := byte(1)
	switch version {
	case 1:
	case 8, 9:
		if count, err = reader.byte(); err != nil {
			return nil, 0, false, 0, err
		}
		if version == 9 {
			y, err := reader.byte()
			if err != nil {
				return nil, 0, false, 0, err
			}
			yIndex, hasYIndex = int8(y), true
		}
	default:
		return nil, 0, false, 0, fmt.Errorf("sub-chunk version %d", version)
	}
	if count > maxStorageLayers {
		return nil, 0, false, 0, fmt.Errorf("%d storage layers", count)
	}
	for i := byte(0); i < count; i++ {
		layer, err := readLayer(reader)
		if err != nil {
			return nil, 0, false, 0, fmt.Errorf("layer %d: %w", i, err)
		}
		layers = append(layers, layer)
	}
	return layers, yIndex, hasYIndex, reader.offset, nil
}

// readLayer reads 4096 palette indices of "bits" bits each, packed into
// little-endian 32-bit words (a block never spans two words), then the palette.
func readLayer(reader *byteReader) (packedLayer, error) {
	var layer packedLayer
	header, err := reader.byte()
	if err != nil {
		return layer, err
	}
	if header&1 == 0 {
		return layer, fmt.Errorf("header %#x: the palette is not block runtime IDs", header)
	}
	// The bit widths the game packs with; 0 bits is one block for the whole layer.
	bits := uint(header >> 1)
	switch bits {
	case 0, 1, 2, 3, 4, 5, 6, 8, 16:
	default:
		return layer, fmt.Errorf("header %#x: %d bits per block", header, bits)
	}
	var words []uint32
	perWord := 0
	if bits > 0 {
		perWord = 32 / int(bits)
		words = make([]uint32, (4096+perWord-1)/perWord)
		for i := range words {
			if words[i], err = reader.uint32LittleEndian(); err != nil {
				return layer, err
			}
		}
	}
	// With 0 bits there is no size: the palette is the one block.
	size := int32(1)
	if bits > 0 {
		if size, err = reader.zigzagVarint32(); err != nil {
			return layer, err
		}
		if size <= 0 || size > 4096 {
			return layer, fmt.Errorf("palette size %d", size)
		}
	}
	layer.palette = make([]uint32, size)
	for i := range layer.palette {
		value, err := reader.zigzagVarint32()
		if err != nil {
			return layer, err
		}
		layer.palette[i] = uint32(value)
	}
	// One palette entry needs no indices: the whole layer is that one.
	if size == 1 {
		return layer, nil
	}
	mask := uint32(1)<<bits - 1
	layer.indices = make([]uint16, 4096)
	for i := range layer.indices {
		index := uint16(words[i/perWord] >> (uint(i%perWord) * bits) & mask)
		if int(index) >= len(layer.palette) {
			return layer, fmt.Errorf("block %d points at palette entry %d of %d", i, index, len(layer.palette))
		}
		layer.indices[i] = index
	}
	return layer, nil
}

// readBiomes reads the biomes of a request-mode level_chunk: one layer per
// sub-chunk from the dimension's lowest up, as a block layer is packed but
// with biome numbers, where a header of 0xff means "the same as the one below".
func readBiomes(data []byte, count int) ([]packedLayer, error) {
	reader := &byteReader{data: data}
	var layers []packedLayer
	for i := 0; i < count; i++ {
		if reader.offset < len(reader.data) && reader.data[reader.offset] == 0xff {
			reader.offset++
			if len(layers) == 0 {
				return nil, errors.New("the lowest biome layer refers to the one below")
			}
			layers = append(layers, layers[len(layers)-1])
			continue
		}
		layer, err := readLayer(reader)
		if err != nil {
			return nil, fmt.Errorf("biome layer %d: %w", i, err)
		}
		layers = append(layers, layer)
	}
	return layers, nil
}

type byteReader struct {
	data   []byte
	offset int
}

var errDataEndedEarly = errors.New("sub-chunk data ended early")

func (r *byteReader) byte() (byte, error) {
	if r.offset >= len(r.data) {
		return 0, errDataEndedEarly
	}
	r.offset++
	return r.data[r.offset-1], nil
}

func (r *byteReader) uint32LittleEndian() (uint32, error) {
	if r.offset+4 > len(r.data) {
		return 0, errDataEndedEarly
	}
	b := r.data[r.offset : r.offset+4]
	r.offset += 4
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24, nil
}

// zigzagVarint32 reads a signed varint32 with the sign in the lowest bit.
func (r *byteReader) zigzagVarint32() (int32, error) {
	var u uint32
	for shift := uint(0); shift < 35; shift += 7 {
		b, err := r.byte()
		if err != nil {
			return 0, err
		}
		u |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			return int32(u>>1) ^ -int32(u&1), nil
		}
	}
	return 0, errors.New("varint too long")
}
