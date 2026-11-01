// Checks that the bot's own health, hunger and effects are written to self,
// and that those of other entities are not.
package arrive

import (
	"testing"

	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const botRuntimeID = 7

func attribute(name string, value, maximum float32) protocol.Attribute {
	return protocol.Attribute{AttributeValue: protocol.AttributeValue{Name: name, Value: value, Max: maximum}}
}

func TestHealthAndFoodOfTheBotOnly(t *testing.T) {
	a := &Arrive{self: self.New(botRuntimeID, botRuntimeID, [3]float32{}, 0)}
	a.updateAttributes(&packet.UpdateAttributes{EntityRuntimeID: botRuntimeID, Attributes: []protocol.Attribute{
		attribute("minecraft:health", 14, 20), attribute("minecraft:player.hunger", 17, 20),
	}})
	a.updateAttributes(&packet.UpdateAttributes{EntityRuntimeID: 99, Attributes: []protocol.Attribute{attribute("minecraft:health", 3, 10)}})
	if now, maximum := a.self.Health(); now != 14 || maximum != 20 {
		t.Errorf("health %v of %v, want 14 of 20", now, maximum)
	}
	if food := a.self.Food(); food != 17 {
		t.Errorf("food %v, want 17", food)
	}
	a.setHealth(&packet.SetHealth{Health: 6})
	if now, maximum := a.self.Health(); now != 6 || maximum != 20 {
		t.Errorf("after set_health: %v of %v, want 6 of 20", now, maximum)
	}
}

func TestEffectsOfTheBotOnly(t *testing.T) {
	a := &Arrive{self: self.New(botRuntimeID, botRuntimeID, [3]float32{}, 0)}
	a.mobEffect(&packet.MobEffect{EntityRuntimeID: botRuntimeID, Operation: packet.MobEffectAdd, EffectType: packet.EffectHaste, Amplifier: 1, Duration: 600})
	a.mobEffect(&packet.MobEffect{EntityRuntimeID: 99, Operation: packet.MobEffectAdd, EffectType: packet.EffectPoison})
	effects := a.self.Effects()
	if len(effects) != 1 || effects[0] != (self.Effect{ID: packet.EffectHaste, Amplifier: 1, Duration: 600}) {
		t.Fatalf("got %+v, want haste 1 for 600 ticks only", effects)
	}
	a.mobEffect(&packet.MobEffect{EntityRuntimeID: botRuntimeID, Operation: packet.MobEffectRemove, EffectType: packet.EffectHaste})
	if effects := a.self.Effects(); len(effects) != 0 {
		t.Errorf("after removing: %+v, want none", effects)
	}
}
