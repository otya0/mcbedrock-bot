// Checks that what is added for the next tick is given once.
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
