// A window opened: a chest, the bag, a crafting table and the like.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) containerOpen(opened *packet.ContainerOpen) {
	a.contents.SetOpenWindow(contents.Window{
		WindowID:       opened.WindowID,
		ContainerType:  opened.ContainerType,
		Position:       opened.ContainerPosition,
		EntityUniqueID: opened.ContainerEntityUniqueID,
	})
}
