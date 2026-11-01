// The attributes of an entity changed (only the changed ones come): for the
// bot itself, write its health and its hunger bar.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) updateAttributes(changed *packet.UpdateAttributes) {
	if changed.EntityRuntimeID != a.self.EntityRuntimeID() {
		return
	}
	for _, attribute := range changed.Attributes {
		switch attribute.Name {
		case "minecraft:health":
			a.self.SetHealth(attribute.Value, attribute.Max)
		case "minecraft:player.hunger":
			a.self.SetFood(attribute.Value)
		}
	}
}
