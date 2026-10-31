// The answer to an item_stack_request. For an accepted one, every slot it lists
// gets its new count and stack ID; the item comes from what the request
// expected, because the answer does not carry it. A refused one changed nothing.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) itemStackResponse(answer *packet.ItemStackResponse) {
	open, _ := a.contents.OpenWindow()
	for _, response := range answer.Responses {
		expected, ok := a.contents.TakeExpected(response.RequestID)
		if !ok || response.Status != protocol.ItemStackResponseStatusOK {
			continue
		}
		for _, info := range response.ContainerInfo {
			for _, slotInfo := range info.SlotInfo {
				item, known := expectedItem(expected, info.Container.ContainerID, slotInfo.Slot)
				windowID, slot, onWindow := contents.WindowOf(open, info.Container.ContainerID, slotInfo.Slot)
				if !known || !onWindow {
					continue
				}
				if slotInfo.Count == 0 {
					item = protocol.ItemInstance{}
				} else {
					item.Stack.Count = uint16(slotInfo.Count)
					item.StackNetworkID = slotInfo.StackNetworkID
				}
				a.contents.SetSlot(windowID, slot, item)
				a.showHand(windowID, slot)
			}
		}
	}
}

func expectedItem(expected []contents.Expected, containerID, slot byte) (protocol.ItemInstance, bool) {
	for _, one := range expected {
		if one.ContainerID == containerID && one.Slot == slot {
			return one.Item, true
		}
	}
	return protocol.ItemInstance{}, false
}
