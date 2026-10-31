// Checks that a key that went down or up is said once, in that tick.
package move

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestChangeFlagsSayStartAndStopOnce(t *testing.T) {
	nothing := keys{}
	sprinting := keys{sprint: true}
	if flags := changeFlags(nothing, sprinting); len(flags) != 1 || flags[0] != packet.InputFlagStartSprinting {
		t.Fatalf("sprint down gave %v, want StartSprinting", flags)
	}
	if flags := changeFlags(sprinting, sprinting); len(flags) != 0 {
		t.Fatalf("sprint still down gave %v, want nothing", flags)
	}
	if flags := changeFlags(sprinting, nothing); len(flags) != 1 || flags[0] != packet.InputFlagStopSprinting {
		t.Fatalf("sprint up gave %v, want StopSprinting", flags)
	}
}

func TestStickFollowsTheKeys(t *testing.T) {
	if direction := stick(keys{forward: true, left: true}); direction != [2]float32{1, 1} {
		t.Fatalf("forward and left gave %v, want {1, 1}", direction)
	}
	if direction := stick(keys{forward: true, back: true}); direction != [2]float32{} {
		t.Fatalf("forward and back gave %v, want {0, 0}", direction)
	}
}

// The elytra opened only with jump pressed in the same tick (the live test, 2026-10-31).
func TestStartGlidePressesJumpInTheSameTick(t *testing.T) {
	m := &Move{self: self.New(1, 1, [3]float32{}, 0), logger: slog.New(slog.DiscardHandler)}
	m.StartGlide()
	flags, _ := m.self.TakeTickInput()
	for _, want := range []int{packet.InputFlagJumping, packet.InputFlagStartJumping, packet.InputFlagStartGliding} {
		if !slices.Contains(flags, want) {
			t.Fatalf("StartGlide gave %v, want flag %d in it", flags, want)
		}
	}
}
