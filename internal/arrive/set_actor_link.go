// The bot got on or off something: write what it rides. The runtime ID of the
// vehicle comes from the entities record; when the vehicle is not known there,
// it is taken to be the same number as the unique ID (a guess).
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) setActorLink(changed *packet.SetActorLink) {
	if changed.EntityLink.RiderEntityUniqueID != a.self.EntityUniqueID() {
		return
	}
	if changed.EntityLink.Type == protocol.EntityLinkRemove {
		a.self.SetVehicle(self.Vehicle{}, false)
		return
	}
	ridden := changed.EntityLink.RiddenEntityUniqueID
	vehicle := self.Vehicle{EntityUniqueID: ridden, EntityRuntimeID: uint64(ridden)}
	if known, ok := a.entities.ByUniqueID(ridden); ok {
		vehicle.EntityRuntimeID = known.RuntimeID
	}
	a.self.SetVehicle(vehicle, true)
}
