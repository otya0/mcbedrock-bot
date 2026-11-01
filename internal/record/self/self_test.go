// Checks that what is added for the next tick is given once, and that the
// effects are kept by ID.
package self

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// What is added for the next tick is given once, then forgotten.
func TestTakeTickInput(t *testing.T) {
	s := New(1, 1, [3]float32{}, 0)
	s.AddTickFlag(packet.InputFlagMissedSwing)
	s.AddTickBlockAction(protocol.PlayerBlockAction{Action: protocol.PlayerActionStartBreak})
	flags, actions := s.TakeTickInput()
	if len(flags) != 1 || len(actions) != 1 {
		t.Fatalf("got %d flags and %d actions, want 1 and 1", len(flags), len(actions))
	}
	flags, actions = s.TakeTickInput()
	if len(flags) != 0 || len(actions) != 0 {
		t.Fatalf("got %d flags and %d actions the second time, want none", len(flags), len(actions))
	}
}

// An effect is changed by its ID, removed by its ID, and given back in ID order.
func TestEffectsByID(t *testing.T) {
	s := New(1, 1, [3]float32{}, 0)
	s.SetEffect(Effect{ID: packet.EffectPoison, Amplifier: 0, Duration: 100})
	s.SetEffect(Effect{ID: packet.EffectSpeed, Amplifier: 0, Duration: 200})
	s.SetEffect(Effect{ID: packet.EffectPoison, Amplifier: 1, Duration: 50})
	effects := s.Effects()
	if len(effects) != 2 || effects[0].ID != packet.EffectSpeed || effects[1].Amplifier != 1 {
		t.Fatalf("got %+v, want speed, then poison with amplifier 1", effects)
	}
	s.RemoveEffect(packet.EffectSpeed)
	if effects := s.Effects(); len(effects) != 1 || effects[0].ID != packet.EffectPoison {
		t.Errorf("after removing speed got %+v, want poison only", effects)
	}
}

// No more than maxEffects distinct effects are kept, whatever the server sends.
func TestEffectsHaveALimit(t *testing.T) {
	s := New(1, 1, [3]float32{}, 0)
	for id := int32(1); id <= maxEffects+10; id++ {
		s.SetEffect(Effect{ID: id})
	}
	if n := len(s.Effects()); n != maxEffects {
		t.Errorf("kept %d effects, want %d", n, maxEffects)
	}
}
