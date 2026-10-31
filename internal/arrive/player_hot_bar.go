// The server says which hotbar slot is selected (once while joining).
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) playerHotBar(hotBar *packet.PlayerHotBar) {
	a.self.SetHotBarSlot(int(hotBar.SelectedHotBarSlot))
}
