// The recipes arrived (while joining).
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) craftingData(recipes *packet.CraftingData) {
	a.contents.SetRecipes(recipes)
}
