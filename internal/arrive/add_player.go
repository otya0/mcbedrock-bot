// Another player appeared: keep it, with its name.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/entities"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) addPlayer(added *packet.AddPlayer) {
	a.entities.Add(entities.Entity{RuntimeID: added.EntityRuntimeID, UniqueID: added.AbilityData.EntityUniqueID, Type: "minecraft:player", Name: added.Username, Position: added.Position})
}
