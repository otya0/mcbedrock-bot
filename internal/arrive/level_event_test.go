// Checks that the cracking speed of the held block becomes its breaking time,
// and that other events and other blocks are left alone.
package arrive

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestCrackingSpeedBecomesBreakingTicks(t *testing.T) {
	a := &Arrive{self: self.New(botRuntimeID, botRuntimeID, [3]float32{}, 0)}
	held := protocol.BlockPos{-84, 62, -94}
	a.self.SetBreaking(self.Breaking{Position: held, Face: 1}, true)
	at := mgl32.Vec3{-84, 62, -94}

	// 3640 is what the server sent for a grass block broken by hand (0.9 seconds).
	a.levelEvent(&packet.LevelEvent{EventType: packet.LevelEventStartBlockCracking, Position: at, EventData: 3640})
	if breaking, _ := a.self.Breaking(); breaking.Ticks != 18 {
		t.Fatalf("ticks %d, want 18", breaking.Ticks)
	}
	// Another block, another kind of event, and a speed of 0 change nothing.
	a.levelEvent(&packet.LevelEvent{EventType: packet.LevelEventStartBlockCracking, Position: mgl32.Vec3{0, 0, 0}, EventData: 100})
	a.levelEvent(&packet.LevelEvent{EventType: packet.LevelEventParticlesDestroyBlock, Position: at, EventData: 100})
	a.levelEvent(&packet.LevelEvent{EventType: packet.LevelEventUpdateBlockCracking, Position: at, EventData: 0})
	if breaking, _ := a.self.Breaking(); breaking.Ticks != 18 {
		t.Fatalf("ticks %d after the other events, want still 18", breaking.Ticks)
	}
	// A new speed (haste, say) changes the time.
	a.levelEvent(&packet.LevelEvent{EventType: packet.LevelEventUpdateBlockCracking, Position: at, EventData: 6553})
	if breaking, _ := a.self.Breaking(); breaking.Ticks != 10 {
		t.Errorf("ticks %d after the update, want 10", breaking.Ticks)
	}
}
