// Hearing the server and the clock: a program can be told every packet that
// arrives (OnPacket, or Options.OnPacket from the first packet on) and every
// tick (OnTick). They run inside the bot, so they must return quickly, and a
// panic in them ends the connection (Err says why). Packets and ticks come in
// two goroutines at the same time: data that both touch needs a mutex or a channel.
package mcbedrock

import (
	"slices"

	"github.com/otya0/mcbedrock-bot/internal/act/move"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// OnPacket adds a function told every packet that arrives from now on, in
// order, after the records have it.
func (b *Bot) OnPacket(observer func(packet.Packet)) {
	b.link.Observe(observer)
}

// OnTick adds a function called once every game tick (20 a second), just
// before the bot sends the tick's input: what it asks of the doors goes out
// in the same tick.
func (b *Bot) OnTick(observer func(tick uint64)) {
	b.tickMu.Lock()
	defer b.tickMu.Unlock()
	// A new slice each time, so that tellTick can go through the one it took
	// without holding the lock.
	b.tickObservers = append(slices.Clone(b.tickObservers), observer)
}

// tick is the clock's work each tick: the OnTick functions first, then the
// tick's input. move.Tick is not a user's job: it is not a method of Move.
func (b *Bot) tick(tick uint64) {
	b.tellTick(tick)
	move.Tick(b.move, tick)
}

func (b *Bot) tellTick(tick uint64) {
	b.tickMu.Lock()
	observers := b.tickObservers
	b.tickMu.Unlock()
	for _, observer := range observers {
		observer(tick)
	}
}
