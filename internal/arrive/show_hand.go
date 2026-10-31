// The item in a hand may have changed: tell the server, as a client does by
// itself, so that others see the hand. Four handlers call this.
package arrive

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// wholeWindow is the slot that means every slot of the window changed.
const wholeWindow = -1

// showHand sends mob_equipment for the main hand when its slot changed, and
// for the off hand when that window changed.
func (a *Arrive) showHand(windowID byte, slot int) {
	switch windowID {
	case protocol.WindowIDInventory:
		selected := a.self.HotBarSlot()
		if slot != wholeWindow && slot != selected {
			return
		}
		_ = a.link.SendMobEquipment(&packet.MobEquipment{
			EntityRuntimeID: a.self.EntityRuntimeID(),
			NewItem:         a.contents.ItemIn(protocol.WindowIDInventory, selected),
			InventorySlot:   byte(selected),
			HotBarSlot:      byte(selected),
			WindowID:        protocol.WindowIDInventory,
		})
	case protocol.WindowIDOffHand:
		// The off hand is slot 1 in mob_equipment, as a client sends it.
		_ = a.link.SendMobEquipment(&packet.MobEquipment{
			EntityRuntimeID: a.self.EntityRuntimeID(),
			NewItem:         a.contents.ItemIn(protocol.WindowIDOffHand, 0),
			InventorySlot:   1,
			HotBarSlot:      1,
			WindowID:        protocol.WindowIDOffHand,
		})
	}
}
