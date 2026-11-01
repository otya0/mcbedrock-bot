// An entity moved, with only the coordinates that changed (each one is the
// new value, not a difference): write where it is now.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) moveActorDelta(moved *packet.MoveActorDelta) {
	entity, known := a.entities.Get(moved.EntityRuntimeID)
	if !known {
		return
	}
	position := entity.Position
	if x, ok := moved.PositionX.Value(); ok {
		position[0] = x
	}
	if y, ok := moved.PositionY.Value(); ok {
		position[1] = y
	}
	if z, ok := moved.PositionZ.Value(); ok {
		position[2] = z
	}
	a.entities.SetPosition(moved.EntityRuntimeID, position)
}
