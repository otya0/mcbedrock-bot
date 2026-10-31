// Checks that a sub-chunk and the biomes in the network form are read.
package arrive

import "testing"

func TestReadSubChunkReadsOnePaletteEntry(t *testing.T) {
	// Version 9, one layer, y index 4; the layer: header 0x01 (0 bits per block),
	// no indices, no size, one palette entry (zigzag varint of 10 is 20).
	data := []byte{9, 1, 4, 0x01, 20}
	layers, yIndex, hasYIndex, used, err := readSubChunk(data)
	if err != nil {
		t.Fatal(err)
	}
	if !hasYIndex || yIndex != 4 || used != len(data) || len(layers) != 1 || layers[0].palette[0] != 10 || layers[0].indices != nil {
		t.Fatalf("got layers %v, y %d (%v), used %d", layers, yIndex, hasYIndex, used)
	}
}

func TestReadBiomesCopiesTheLayerBelow(t *testing.T) {
	// Biome 1 (zigzag 2) for the lowest sub-chunk, then 0xff: the same again.
	layers, err := readBiomes([]byte{0x01, 2, 0xff}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(layers) != 2 || layers[0].palette[0] != 1 || layers[1].palette[0] != 1 {
		t.Fatalf("got %v", layers)
	}
	if _, err := readBiomes([]byte{0xff}, 1); err == nil {
		t.Fatal("no error when the lowest layer refers to the one below")
	}
}
