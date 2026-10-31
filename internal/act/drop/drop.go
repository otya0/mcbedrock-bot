// Package drop is Q: with no window open, one of the held item falls
// (inventory_transaction); with a window open, what the cursor holds falls
// (item_stack_request).
package drop

import (
	"errors"
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Drop struct {
	link     *serverlink.Link
	self     *self.Self
	contents *contents.Contents
	logger   *slog.Logger
}

func New(link *serverlink.Link, selfRecord *self.Self, contentsRecord *contents.Contents, logger *slog.Logger) *Drop {
	return &Drop{link: link, self: selfRecord, contents: contentsRecord, logger: logger}
}

func (d *Drop) HeldItem() error {
	if _, open := d.contents.OpenWindow(); open {
		return d.cursor()
	}
	slot := d.self.HotBarSlot()
	held := d.contents.ItemIn(protocol.WindowIDInventory, slot)
	if contents.IsEmpty(held) {
		return errors.New("drop held item: the hand is empty")
	}
	left := held
	left.Stack.Count--
	if left.Stack.Count == 0 {
		left = protocol.ItemInstance{}
	}
	dropped := held
	dropped.Stack.Count = 1
	d.logger.Info("act", "func", "Drop.HeldItem", "slot", slot, "packet", "inventory_transaction/Normal")
	err := d.link.SendInventoryTransaction(&packet.InventoryTransaction{
		LegacyRequestID:    d.contents.NextRequestID(),
		LegacySetItemSlots: []protocol.LegacySetItemSlot{{ContainerID: protocol.ContainerHotBar, Slots: []byte{byte(slot)}}},
		Actions: []protocol.InventoryAction{
			{SourceType: protocol.InventoryActionSourceContainer, WindowID: protocol.Option(int8(protocol.WindowIDInventory)), InventorySlot: uint32(slot), OldItem: held, NewItem: left},
			{SourceType: protocol.InventoryActionSourceWorld, SourceFlags: protocol.Option(uint32(0)), NewItem: dropped},
		},
		TransactionData: &protocol.NormalTransactionData{},
	})
	if err != nil {
		return err
	}
	// The server answers nothing when the drop succeeds, so the slot changes here.
	d.contents.SetSlot(protocol.WindowIDInventory, slot, left)
	return nil
}

// cursor drops all the cursor holds, as a click outside the slots does.
func (d *Drop) cursor() error {
	onCursor := d.contents.ItemIn(protocol.WindowIDUI, 0)
	if contents.IsEmpty(onCursor) {
		return errors.New("drop held item: the cursor holds nothing")
	}
	requestID := d.contents.NextRequestID()
	d.contents.SetExpected(requestID, []contents.Expected{{ContainerID: protocol.ContainerCursor, Slot: 0, Item: onCursor}})
	d.logger.Info("act", "func", "Drop.HeldItem", "window", protocol.WindowIDUI, "slot", 0, "packet", "item_stack_request/Drop")
	return d.link.SendItemStackRequest(&packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{
		RequestID: requestID,
		Actions: []protocol.StackRequestAction{&protocol.DropStackRequestAction{
			Count:  byte(onCursor.Stack.Count),
			Source: protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: protocol.ContainerCursor}, StackNetworkID: onCursor.StackNetworkID},
		}},
	}}})
}
