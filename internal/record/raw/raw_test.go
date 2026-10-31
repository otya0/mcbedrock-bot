// Checks that raw keeps the last 100 of a kind and counts every arrival.
package raw

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestRawKeepsTheLast100(t *testing.T) {
	r := New()
	for i := 1; i <= 101; i++ {
		r.Add(&packet.SetHealth{Health: int32(i)})
	}
	last := r.Last("SetHealth")
	if len(last) != 100 {
		t.Fatalf("kept %d, want 100", len(last))
	}
	oldest, newest := last[0].(*packet.SetHealth).Health, last[99].(*packet.SetHealth).Health
	if oldest != 2 || newest != 101 {
		t.Errorf("kept %d..%d, want 2..101", oldest, newest)
	}
	if n := r.Count()["SetHealth"]; n != 101 {
		t.Errorf("counted %d, want 101", n)
	}
	if len(r.Last("SetTime")) != 0 {
		t.Error("a kind that never arrived should be empty")
	}
}
