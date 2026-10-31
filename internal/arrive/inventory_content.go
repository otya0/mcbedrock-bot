// All slots of one window arrived.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) inventoryContent(content *packet.InventoryContent) {
	a.contents.SetWindowItems(byte(content.WindowID), content.Content)
	a.showHand(byte(content.WindowID), wholeWindow)
}
