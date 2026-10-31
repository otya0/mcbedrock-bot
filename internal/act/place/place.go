// Package place puts the held block on a face of a block (the right button).
package place

import (
	"fmt"
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Place struct {
	link     *serverlink.Link
	self     *self.Self
	contents *contents.Contents
	blocks   *blocks.Blocks
	logger   *slog.Logger
}

func New(link *serverlink.Link, selfRecord *self.Self, contentsRecord *contents.Contents, blocksRecord *blocks.Blocks, logger *slog.Logger) *Place {
	return &Place{link: link, self: selfRecord, contents: contentsRecord, blocks: blocksRecord, logger: logger}
}

// OnBlock right-clicks a face of a block with the held item; the new block goes
// on the other side of that face. The clicked block's number is sent as 0 when
// the bot does not have it.
func (p *Place) OnBlock(position protocol.BlockPos, face int32) error {
	if !blocks.IsFace(face) {
		return fmt.Errorf("place on block: face %d is not 0 to 5", face)
	}
	slot := p.self.HotBarSlot()
	target, _ := p.blocks.At(position)
	p.logger.Info("act", "func", "Place.OnBlock", "target", blocks.PositionText(position), "face", face, "packet", "inventory_transaction/UseItem/ClickBlock")
	return p.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.UseItemTransactionData{
		ActionType:       protocol.UseItemActionClickBlock,
		TriggerType:      protocol.TriggerTypePlayerInput,
		BlockPosition:    position,
		BlockFace:        face,
		HotBarSlot:       int32(slot),
		Hand:             protocol.HandSlotMainHand,
		HeldItem:         p.contents.ItemIn(protocol.WindowIDInventory, slot),
		Position:         p.self.Position(),
		ClickedPosition:  blocks.FaceCentre(face),
		BlockRuntimeID:   target.RuntimeID,
		ClientPrediction: protocol.ClientPredictionSuccess,
	}})
}
