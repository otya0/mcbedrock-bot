// A horse's (or another mount's) window opened.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) updateEquip(equip *packet.UpdateEquip) {
	a.contents.SetOpenWindow(contents.Window{WindowID: equip.WindowID, ContainerType: equip.WindowType, EntityUniqueID: equip.EntityUniqueID})
}
