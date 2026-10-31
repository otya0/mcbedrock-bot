// Package raw keeps every packet that arrives: the last 100 of each kind, as
// they arrived, and how many of each kind came.
package raw

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const keep = 100

type Raw struct {
	mu    sync.Mutex
	last  map[string][]packet.Packet
	count map[string]int
}

func New() *Raw {
	return &Raw{last: map[string][]packet.Packet{}, count: map[string]int{}}
}

// Add keeps a packet under its kind: the packet's type name, "SetHealth".
func (r *Raw) Add(arrived packet.Packet) {
	kind := strings.TrimPrefix(fmt.Sprintf("%T", arrived), "*packet.")
	r.mu.Lock()
	defer r.mu.Unlock()
	r.count[kind]++
	kept := append(r.last[kind], arrived)
	if len(kept) > keep {
		kept = kept[1:]
	}
	r.last[kind] = kept
}

// Last gives the kept packets of one kind, oldest first.
func (r *Raw) Last(kind string) []packet.Packet {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]packet.Packet(nil), r.last[kind]...)
}

// Count gives how many packets of each kind have arrived since joining.
func (r *Raw) Count() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := map[string]int{}
	for kind, n := range r.count {
		count[kind] = n
	}
	return count
}
