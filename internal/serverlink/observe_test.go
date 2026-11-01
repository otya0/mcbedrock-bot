// Checks that observers hear every packet after its listener, the ones that
// arrived while entering included, and that a panic in one ends the connection.
package serverlink

import (
	"strings"
	"sync"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestObserverIsToldAfterTheListener(t *testing.T) {
	link := linkForTest()
	var order []string
	link.Terrain.UpdateBlock = func(*packet.UpdateBlock) { order = append(order, "listener") }
	link.Observe(func(packet.Packet) { order = append(order, "observer 1") })
	link.Observe(func(packet.Packet) { order = append(order, "observer 2") })
	link.deliver(&packet.UpdateBlock{})
	if strings.Join(order, ", ") != "listener, observer 1, observer 2" {
		t.Fatalf("told in the order %v", order)
	}
}

// The packets that arrived while entering go first, in order, to an observer
// added before Start. The receive loop starts right after and may add more in
// its own goroutine: hence the mutex, and only the first two are checked.
func TestObserverHearsWhatArrivedWhileEntering(t *testing.T) {
	link := linkForTest()
	entering := []packet.Packet{&packet.UpdateBlock{}, &packet.SetTime{Time: 6000}}
	link.arrivedWhileEntering = entering
	var mu sync.Mutex
	var heard []packet.Packet
	link.Observe(func(arrived packet.Packet) {
		mu.Lock()
		defer mu.Unlock()
		heard = append(heard, arrived)
	})
	link.Start()
	mu.Lock()
	defer mu.Unlock()
	if len(heard) < 2 || heard[0] != entering[0] || heard[1] != entering[1] {
		t.Fatalf("heard %v first, want the two packets that arrived while entering", heard)
	}
}

func TestPanicInObserverEndsTheConnection(t *testing.T) {
	link := linkForTest()
	link.Observe(func(packet.Packet) { panic("a bug in an observer") })
	link.arrivedWhileEntering = []packet.Packet{&packet.UpdateBlock{}}
	link.Start()
	select {
	case <-link.Done():
	default:
		t.Fatal("the connection did not end")
	}
	if err := link.Err(); err == nil || !strings.Contains(err.Error(), "a bug in an observer") {
		t.Fatalf("Err is %v, want the panic", err)
	}
}
