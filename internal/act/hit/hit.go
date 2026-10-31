// Package hit is the left button on an entity, and the swing of the arm.
package hit

import (
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Hit struct {
	link     *serverlink.Link
	self     *self.Self
	contents *contents.Contents
	logger   *slog.Logger
}

func New(link *serverlink.Link, selfRecord *self.Self, contentsRecord *contents.Contents, logger *slog.Logger) *Hit {
	return &Hit{link: link, self: selfRecord, contents: contentsRecord, logger: logger}
}

// Entity hits an entity with the held item; whether it gets hurt is the server's.
func (h *Hit) Entity(runtimeID uint64) error {
	slot := h.self.HotBarSlot()
	h.logger.Info("act", "func", "Hit.Entity", "entity_runtime_id", runtimeID, "packet", "inventory_transaction/UseItemOnEntity/Attack")
	return h.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.UseItemOnEntityTransactionData{
		TargetEntityRuntimeID: runtimeID,
		ActionType:            protocol.UseItemOnEntityActionAttack,
		HotBarSlot:            int32(slot),
		HeldItem:              h.contents.ItemIn(protocol.WindowIDInventory, slot),
		Position:              h.self.Position(),
		ClickedPosition:       [3]float32{0, 1, 0},
	}})
}

// Swing shows the arm swinging to others.
func (h *Hit) Swing() error {
	h.logger.Info("act", "func", "Hit.Swing", "packet", "animate/SwingArm")
	return h.link.SendAnimate(&packet.Animate{ActionType: packet.AnimateActionSwingArm, EntityRuntimeID: h.self.EntityRuntimeID()})
}

// Air is a left click on nothing: a flag in the next tick's input.
func (h *Hit) Air() {
	h.self.AddTickFlag(packet.InputFlagMissedSwing)
	h.logger.Info("act", "func", "Hit.Air", "packet", "player_auth_input/MissedSwing")
}
