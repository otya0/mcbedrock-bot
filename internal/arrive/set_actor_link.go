// The bot got on or off something: write what it rides. The runtime ID of the
// vehicle is looked up among the add_actor packets kept in raw; when it is not
// there, it is taken to be the same number as the unique ID (a guess).
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
	for _, kept := range a.raw.Last("AddActor") {
		if added, ok := kept.(*packet.AddActor); ok && added.EntityUniqueID == ridden {
			vehicle.EntityRuntimeID = added.EntityRuntimeID
		}
	}
	a.self.SetVehicle(vehicle, true)
}
