// Package choose selects a hotbar slot (the keys 1 to 9 select slots 0 to 8):
// one mob_equipment with the item of that slot, from which the server shows
// the hand.
package choose

import (
	"fmt"
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Choose struct {
	link     *serverlink.Link
	self     *self.Self
	contents *contents.Contents
	logger   *slog.Logger
}

func New(link *serverlink.Link, selfRecord *self.Self, contentsRecord *contents.Contents, logger *slog.Logger) *Choose {
	return &Choose{link: link, self: selfRecord, contents: contentsRecord, logger: logger}
}

// Slot selects hotbar slot 0 to 8.
func (c *Choose) Slot(slot int) error {
	if slot < 0 || slot > 8 {
		return fmt.Errorf("choose slot: %d is not a hotbar slot (0 to 8)", slot)
	}
	c.self.SetHotBarSlot(slot)
	c.logger.Info("act", "func", "Choose.Slot", "slot", slot, "packet", "mob_equipment")
	return c.link.SendMobEquipment(&packet.MobEquipment{
		EntityRuntimeID: c.self.EntityRuntimeID(),
		NewItem:         c.contents.ItemIn(protocol.WindowIDInventory, slot),
		InventorySlot:   byte(slot),
		HotBarSlot:      byte(slot),
		WindowID:        protocol.WindowIDInventory,
	})
}
