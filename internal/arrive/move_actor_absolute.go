// An entity moved, with its whole position: write where it is now.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) moveActorAbsolute(moved *packet.MoveActorAbsolute) {
	a.entities.SetPosition(moved.EntityRuntimeID, moved.Position)
}
