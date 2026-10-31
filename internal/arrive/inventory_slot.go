// One slot of a window changed.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) inventorySlot(changed *packet.InventorySlot) {
	a.contents.SetSlot(byte(changed.WindowID), int(changed.Slot), changed.NewItem)
	a.showHand(byte(changed.WindowID), int(changed.Slot))
}
