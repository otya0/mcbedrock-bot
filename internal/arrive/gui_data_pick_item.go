// After a middle click the server says which hotbar slot to select.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) guiDataPickItem(picked *packet.GUIDataPickItem) {
	a.self.SetHotBarSlot(int(picked.HotBarSlot))
}
