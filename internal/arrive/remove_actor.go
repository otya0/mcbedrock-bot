// An entity disappeared (died, went out of range, was picked up): forget it.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) removeActor(removed *packet.RemoveActor) {
	a.entities.RemoveByUniqueID(removed.EntityUniqueID)
}
