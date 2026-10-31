// Biome names: the game's own biomes from Mojang's list, and the world's own
// from the server's biome list. The server's list gives the game's biomes the
// number -1 (2026-10-22: 89 biomes, all -1), so only the ones it numbers are taken.
package names

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// SetWorldBiomes takes the biomes the world adds (biome_definition_list).
func (n *Names) SetWorldBiomes(list *packet.BiomeDefinitionList) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, definition := range list.BiomeDefinitions {
		// A name index outside the string list would stop the bot: skip it.
		if definition.BiomeID < 0 || definition.NameIndex < 0 || int(definition.NameIndex) >= len(list.StringList) {
			continue
		}
		n.biomes[int32(definition.BiomeID)] = list.StringList[definition.NameIndex]
	}
}

// Biome is the name of a biome number ("minecraft:plains"), "" if not listed.
func (n *Names) Biome(id uint32) string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.biomes[int32(id)]
}
