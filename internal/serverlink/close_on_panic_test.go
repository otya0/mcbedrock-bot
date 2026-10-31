// Checks that a panic in a listener ends the connection instead of the program.
package serverlink

import (
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestPanicInListenerEndsTheConnection(t *testing.T) {
	link := linkForTest()
	link.Terrain.UpdateBlock = func(*packet.UpdateBlock) { panic("a bug in a listener") }
	link.arrivedWhileEntering = []packet.Packet{&packet.UpdateBlock{}}
	link.Start()
	select {
	case <-link.Done():
	default:
		t.Fatal("the connection did not end")
	}
	if err := link.Err(); err == nil || !strings.Contains(err.Error(), "a bug in a listener") {
		t.Fatalf("Err is %v, want the panic", err)
	}
}
