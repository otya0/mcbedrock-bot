// A level event. The ones read here say how fast the block the bot breaks
// cracks: when breaking starts, and again when the speed changes. The speed is
// 65535 divided by the ticks the whole break takes; write those ticks.
package arrive

import (
	"math"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) levelEvent(event *packet.LevelEvent) {
	if event.EventType != packet.LevelEventStartBlockCracking && event.EventType != packet.LevelEventUpdateBlockCracking {
		return
	}
	if event.EventData <= 0 {
		return
	}
	position := protocol.BlockPos{
		int32(math.Floor(float64(event.Position[0]))),
		int32(math.Floor(float64(event.Position[1]))),
		int32(math.Floor(float64(event.Position[2]))),
	}
	a.self.SetBreakingTicks(position, 65535/int(event.EventData))
}
