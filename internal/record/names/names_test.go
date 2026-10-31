// Checks that Mojang's data is read so that block and biome numbers come out
// as the server sends them. Mojang's files are not part of mcbedrock-bot: the first
// test runs only where download-mojang-data has saved them.
package names

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestNamesFromMojangData(t *testing.T) {
	folder, err := MojangDataFolder()
	if err != nil {
		t.Skip(err)
	}
	n, err := New(folder)
	if err != nil {
		t.Skipf("%v (run download-mojang-data first)", err)
	}
	// 3690217760 is air, as the server sends it (record/blocks.AirRuntimeID).
	if name, _ := n.Block(3690217760); name != "minecraft:air" {
		t.Fatalf("block 3690217760 is %q, want minecraft:air", name)
	}
	if len(n.blocks) < 10000 {
		t.Fatalf("only %d block states", len(n.blocks))
	}
	if name := n.Biome(1); name != "minecraft:plains" {
		t.Fatalf("biome 1 is %q, want minecraft:plains", name)
	}
}

func TestMissingMojangDataSaysHowToTakeIt(t *testing.T) {
	_, err := New(t.TempDir())
	if err == nil {
		t.Fatal("no error for an empty folder")
	}
	t.Log(err)
}

// A world biome whose name index is outside the string list is skipped, not
// read out of range.
func TestWorldBiomesSkipBadNameIndex(t *testing.T) {
	n := &Names{biomes: map[int32]string{}}
	n.SetWorldBiomes(&packet.BiomeDefinitionList{
		StringList: []string{"mcbedrock:test"},
		BiomeDefinitions: []protocol.BiomeDefinition{
			{NameIndex: -1, BiomeID: 200},
			{NameIndex: 5, BiomeID: 201},
			{NameIndex: 0, BiomeID: 202},
		},
	})
	if n.Biome(200) != "" || n.Biome(201) != "" || n.Biome(202) != "mcbedrock:test" {
		t.Errorf("got %q %q %q", n.Biome(200), n.Biome(201), n.Biome(202))
	}
}
