// Checks that At gives the block and biome of a position, and that SetBlock
// changes one block of a layer that was all one block.
package blocks

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestAtAndSetBlock(t *testing.T) {
	b := New()
	b.SetListed([2]int32{0, 0}, -1, true)
	b.SetSubChunk([3]int32{0, 0, 0}, SubChunk{Layers: []Layer{{Palette: []Block{Air()}}}})
	b.SetBiomes([2]int32{0, 0}, 0, []BiomeLayer{{Palette: []string{"minecraft:plains"}}})

	position := protocol.BlockPos{1, 2, 3}
	if block, ok := b.At(position); !ok || block.Name != "minecraft:air" || block.Biome != "minecraft:plains" {
		t.Fatalf("At gave %+v, %v; want air in plains", block, ok)
	}
	dirt := Block{RuntimeID: 2, Name: "minecraft:dirt"}
	b.SetBlock(position, 0, dirt)
	if block, _ := b.At(position); block.Name != "minecraft:dirt" {
		t.Fatalf("after SetBlock, At gave %+v; want dirt", block)
	}
	if block, _ := b.At(protocol.BlockPos{1, 2, 4}); block.Name != "minecraft:air" {
		t.Fatalf("the block next to it became %+v; want air", block)
	}
	if _, ok := b.At(protocol.BlockPos{100, 2, 3}); ok {
		t.Fatal("At answered for a sub-chunk the bot does not have")
	}
}
