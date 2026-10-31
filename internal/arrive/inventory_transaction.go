// The server says which slot a picked-up item went into.
package arrive

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) inventoryTransaction(transaction *packet.InventoryTransaction) {
	for _, action := range transaction.Actions {
		if action.SourceType != protocol.InventoryActionSourceContainer {
			continue
		}
		windowID, ok := action.WindowID.Value()
		if !ok {
			continue
		}
		a.contents.SetSlot(byte(windowID), int(action.InventorySlot), action.NewItem)
		a.showHand(byte(windowID), int(action.InventorySlot))
	}
}
