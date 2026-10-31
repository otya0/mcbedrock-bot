package serverlink

import (
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// A Link with no connection is enough for handOut and the tick clock.
func linkForTest() *Link {
	return &Link{done: make(chan struct{})}
}

func TestHandOutTellsTheListenerOfTheSystem(t *testing.T) {
	link := linkForTest()
	var told *packet.ContainerOpen
	link.Inventory.ContainerOpen = func(arrived *packet.ContainerOpen) { told = arrived }
	opened := &packet.ContainerOpen{WindowID: 7}
	link.handOut(opened)
	if told != opened {
		t.Fatalf("the inventory listener was not told the container_open")
	}
}

func TestHandOutDropsAPacketWithoutAListener(t *testing.T) {
	link := linkForTest()
	link.handOut(&packet.UpdateBlock{}) // must not panic
}

func TestHandOutTellsOtherPacketForAPacketOfNoSystem(t *testing.T) {
	link := linkForTest()
	var told packet.Packet
	link.OtherPacket = func(arrived packet.Packet) { told = arrived }
	request := &packet.CommandRequest{}
	link.handOut(request)
	if told != request {
		t.Fatalf("OtherPacket was not told the command_request")
	}
}

func TestTickClockCountsFromOneAndStopsWhenDone(t *testing.T) {
	link := linkForTest()
	ticks := make(chan uint64, 100)
	link.StartTickClock(func(tickNumber uint64) { ticks <- tickNumber })
	first := <-ticks
	if first != 1 {
		t.Fatalf("the first tick number was %d, want 1", first)
	}
	link.finish(ErrClosed)
	// After done, no new tick may come once the ones already fired drain.
	time.Sleep(3 * TickInterval)
	for len(ticks) > 0 {
		<-ticks
	}
	select {
	case tickNumber := <-ticks:
		t.Fatalf("the clock ticked (%d) after the link was done", tickNumber)
	case <-time.After(3 * TickInterval):
	}
}

func TestFinishKeepsTheFirstReasonAndClosesDoneOnce(t *testing.T) {
	link := linkForTest()
	link.finish(ErrClosed)
	link.finish(nil)
	if link.Err() != ErrClosed {
		t.Fatalf("Err is %v, want ErrClosed", link.Err())
	}
	select {
	case <-link.Done():
	default:
		t.Fatalf("Done is not closed")
	}
}
