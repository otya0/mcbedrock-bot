// An effect on an entity was added, changed or removed: for the bot itself,
// write it.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) mobEffect(changed *packet.MobEffect) {
	if changed.EntityRuntimeID != a.self.EntityRuntimeID() {
		return
	}
	switch changed.Operation {
	case packet.MobEffectAdd, packet.MobEffectModify:
		a.self.SetEffect(self.Effect{ID: changed.EffectType, Amplifier: changed.Amplifier, Duration: changed.Duration})
	case packet.MobEffectRemove:
		a.self.RemoveEffect(changed.EffectType)
	}
}
