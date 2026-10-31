// The world's biome list arrived (while joining): it names the biomes the
// world adds; the game's own are named from Mojang's data.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) biomeDefinitionList(list *packet.BiomeDefinitionList) {
	a.names.SetWorldBiomes(list)
}
