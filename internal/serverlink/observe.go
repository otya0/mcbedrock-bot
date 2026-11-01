// Observers: functions told every packet that arrives, after the listener of
// its system has handled it. The listeners are layer 2's; a program hears the
// server through the observers, without taking a listener's place.
package serverlink

import (
	"slices"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Observe adds a function told every packet from now on, after the listener
// of its system. It can be called at any time; one added before Start is also
// told the packets that arrived while entering. It runs in the receive loop,
// so it must return quickly, and a panic in it ends the connection.
func (link *Link) Observe(observer func(packet.Packet)) {
	link.observersMu.Lock()
	defer link.observersMu.Unlock()
	// A new slice each time, so that deliver can go through the one it took
	// without holding the lock.
	link.observers = append(slices.Clone(link.observers), observer)
}

// deliver gives a packet to the listener of its system, then to every observer.
func (link *Link) deliver(arrived packet.Packet) {
	link.handOut(arrived)
	link.observersMu.Lock()
	observers := link.observers
	link.observersMu.Unlock()
	for _, observer := range observers {
		observer(arrived)
	}
}
