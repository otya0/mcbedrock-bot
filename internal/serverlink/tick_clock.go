package serverlink

import "time"

// TickInterval is the length of one game tick. Minecraft runs at 20 ticks
// a second, and a real client sends player_auth_input once a tick.
const TickInterval = 50 * time.Millisecond

// StartTickClock calls onTick once every TickInterval, in a goroutine of
// its own, until the connection ends. The tick number starts at 1 and
// goes up by one each time.
//
// The clock sends nothing by itself: onTick decides what, if anything, to
// send. Call it once; a second call starts a second clock.
func (link *Link) StartTickClock(onTick func(tickNumber uint64)) {
	go func() {
		defer link.closeOnPanic()
		ticker := time.NewTicker(TickInterval)
		defer ticker.Stop()
		var tickNumber uint64
		for {
			select {
			case <-link.done:
				return
			case <-ticker.C:
			}
			tickNumber++
			onTick(tickNumber)
		}
	}()
}
