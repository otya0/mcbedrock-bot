// Checks that the slot numbers a request uses and the ones an answer uses
// turn into each other.
package contents

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Every slot the container door names must come back as the same window and
// slot from the answer's container ID.
func TestContainerIDOfAndBack(t *testing.T) {
	chest := Window{WindowID: 3, ContainerType: protocol.ContainerTypeContainer}
	furnace := Window{WindowID: 4, ContainerType: protocol.ContainerTypeFurnace}
	blastFurnace := Window{WindowID: 5, ContainerType: protocol.ContainerTypeBlastFurnace}
	brewingStand := Window{WindowID: 6, ContainerType: protocol.ContainerTypeBrewingStand}
	cases := []struct {
		open   Window
		window byte
		slot   int
	}{
		{chest, protocol.WindowIDInventory, 0},
		{chest, protocol.WindowIDInventory, 8},
		{chest, protocol.WindowIDInventory, 9},
		{chest, protocol.WindowIDInventory, 35},
		{chest, protocol.WindowIDArmour, 3},
		{chest, protocol.WindowIDOffHand, 0},
		{chest, 3, 0},
		{chest, 3, 26},
		{furnace, 4, 0},
		{furnace, 4, 1},
		{furnace, 4, 2},
		{blastFurnace, 5, 0},
		{brewingStand, 6, 0},
		{brewingStand, 6, 2},
		{brewingStand, 6, 4},
		{chest, protocol.WindowIDUI, 0},
		{chest, protocol.WindowIDUI, 1},
		{chest, protocol.WindowIDUI, 3},
		{chest, protocol.WindowIDUI, 4},
		{chest, protocol.WindowIDUI, 7},
		{chest, protocol.WindowIDUI, 14},
		{chest, protocol.WindowIDUI, 15},
		{chest, protocol.WindowIDUI, 27},
		{chest, protocol.WindowIDUI, 28},
		{chest, protocol.WindowIDUI, 40},
		{chest, protocol.WindowIDUI, 50},
		{chest, protocol.WindowIDUI, 53},
	}
	for _, c := range cases {
		containerID, ok := ContainerIDOf(c.open, c.window, c.slot)
		if !ok {
			t.Errorf("window %d slot %d: no container ID", c.window, c.slot)
			continue
		}
		window, slot, ok := WindowOf(c.open, containerID, byte(c.slot))
		if !ok || window != c.window || slot != c.slot {
			t.Errorf("window %d slot %d: container %d came back as window %d slot %d (%v)", c.window, c.slot, containerID, window, slot, ok)
		}
	}
}

func TestContainerIDOfRefusesUnknownSlots(t *testing.T) {
	chest := Window{WindowID: 3, ContainerType: protocol.ContainerTypeContainer}
	for _, c := range []struct {
		window byte
		slot   int
	}{{protocol.WindowIDInventory, 36}, {protocol.WindowIDArmour, 4}, {protocol.WindowIDUI, 8}, {protocol.WindowIDUI, 60}, {9, 0}} {
		if _, ok := ContainerIDOf(chest, c.window, c.slot); ok {
			t.Errorf("window %d slot %d: want no container ID", c.window, c.slot)
		}
	}
}
