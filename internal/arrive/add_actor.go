// An entity appeared (an animal, a monster, a boat...): keep it.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/entities"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) addActor(added *packet.AddActor) {
	a.entities.Add(entities.Entity{RuntimeID: added.EntityRuntimeID, UniqueID: added.EntityUniqueID, Type: added.EntityType, Position: added.Position})
}
