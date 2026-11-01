// An item appeared on the ground: keep it, named by what the item is.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/entities"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) addItemActor(added *packet.AddItemActor) {
	name := a.names.Item(added.Item.Stack.ItemType.NetworkID)
	a.entities.Add(entities.Entity{RuntimeID: added.EntityRuntimeID, UniqueID: added.EntityUniqueID, Type: "minecraft:item", Name: name, Position: added.Position})
}
