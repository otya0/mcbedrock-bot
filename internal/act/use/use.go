// Package use is the right button with the held item: on a block whose window
// does not open, on an entity, or in the air (eat, drink, throw, draw a bow).
// What happens is the server's; nothing here depends on the item.
package use

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Use struct {
	link     *serverlink.Link
	self     *self.Self
	contents *contents.Contents
	blocks   *blocks.Blocks
	logger   *slog.Logger

	// The block the right button is held on; only this door needs it.
	mu        sync.Mutex
	usingOn   protocol.BlockPos
	usingFace int32
	holding   bool
}

func New(link *serverlink.Link, selfRecord *self.Self, contentsRecord *contents.Contents, blocksRecord *blocks.Blocks, logger *slog.Logger) *Use {
	return &Use{link: link, self: selfRecord, contents: contentsRecord, blocks: blocksRecord, logger: logger}
}

// OnBlock right-clicks a face of a block: bone meal, flint and steel, a bucket,
// raw meat on a campfire.
func (u *Use) OnBlock(position protocol.BlockPos, face int32) error {
	if !blocks.IsFace(face) {
		return fmt.Errorf("use on block: face %d is not 0 to 5", face)
	}
	slot := u.self.HotBarSlot()
	target, _ := u.blocks.At(position)
	u.logger.Info("act", "func", "Use.OnBlock", "target", blocks.PositionText(position), "face", face, "packet", "inventory_transaction/UseItem/ClickBlock")
	return u.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.UseItemTransactionData{
		ActionType:       protocol.UseItemActionClickBlock,
		TriggerType:      protocol.TriggerTypePlayerInput,
		BlockPosition:    position,
		BlockFace:        face,
		HotBarSlot:       int32(slot),
		Hand:             protocol.HandSlotMainHand,
		HeldItem:         u.contents.ItemIn(protocol.WindowIDInventory, slot),
		Position:         u.self.Position(),
		ClickedPosition:  blocks.FaceCentre(face),
		BlockRuntimeID:   target.RuntimeID,
		ClientPrediction: protocol.ClientPredictionSuccess,
	}})
}

// StartOnBlock holds the right button on a block (a brush); StopOnBlock lets go.
func (u *Use) StartOnBlock(position protocol.BlockPos, face int32) error {
	if !blocks.IsFace(face) {
		return fmt.Errorf("start on block: face %d is not 0 to 5", face)
	}
	u.mu.Lock()
	u.usingOn, u.usingFace, u.holding = position, face, true
	u.mu.Unlock()
	u.logger.Info("act", "func", "Use.StartOnBlock", "target", blocks.PositionText(position), "face", face, "packet", "player_action/StartItemUseOn")
	return u.link.SendPlayerAction(&packet.PlayerAction{EntityRuntimeID: u.self.EntityRuntimeID(), ActionType: protocol.PlayerActionStartItemUseOn, BlockPosition: position, ResultPosition: position, BlockFace: face})
}

func (u *Use) StopOnBlock() error {
	u.mu.Lock()
	position, face, holding := u.usingOn, u.usingFace, u.holding
	u.holding = false
	u.mu.Unlock()
	if !holding {
		return errors.New("stop on block: the right button is not held on a block")
	}
	u.logger.Info("act", "func", "Use.StopOnBlock", "target", blocks.PositionText(position), "packet", "player_action/StopItemUseOn")
	return u.link.SendPlayerAction(&packet.PlayerAction{EntityRuntimeID: u.self.EntityRuntimeID(), ActionType: protocol.PlayerActionStopItemUseOn, BlockPosition: position, ResultPosition: position, BlockFace: face})
}

// OnEntity right-clicks an entity whose window does not open: shear, milk,
// feed, put on a lead, ride.
func (u *Use) OnEntity(runtimeID uint64) error {
	slot := u.self.HotBarSlot()
	u.logger.Info("act", "func", "Use.OnEntity", "entity_runtime_id", runtimeID, "packet", "inventory_transaction/UseItemOnEntity/Interact")
	return u.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.UseItemOnEntityTransactionData{
		TargetEntityRuntimeID: runtimeID,
		ActionType:            protocol.UseItemOnEntityActionInteract,
		HotBarSlot:            int32(slot),
		HeldItem:              u.contents.ItemIn(protocol.WindowIDInventory, slot),
		Position:              u.self.Position(),
		ClickedPosition:       [3]float32{0, 1, 0},
	}})
}

// StartInAir right-clicks with nothing in front: starts eating or drinking,
// throws, draws a bow, puts on armour.
func (u *Use) StartInAir() error {
	slot := u.self.HotBarSlot()
	u.logger.Info("act", "func", "Use.StartInAir", "packet", "inventory_transaction/UseItem/ClickAir")
	return u.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.UseItemTransactionData{
		ActionType:       protocol.UseItemActionClickAir,
		TriggerType:      protocol.TriggerTypePlayerInput,
		HotBarSlot:       int32(slot),
		Hand:             protocol.HandSlotMainHand,
		HeldItem:         u.contents.ItemIn(protocol.WindowIDInventory, slot),
		Position:         u.self.Position(),
		ClientPrediction: protocol.ClientPredictionSuccess,
	}})
}

// StopInAir lets go before the item is used up (shoots the bow).
func (u *Use) StopInAir() error {
	return u.release("Use.StopInAir", "inventory_transaction/ReleaseItem/Release", protocol.ReleaseItemActionRelease)
}

// FinishInAir says the item was used up, as a client does when eating ends.
func (u *Use) FinishInAir() error {
	return u.release("Use.FinishInAir", "inventory_transaction/ReleaseItem/Consume", protocol.ReleaseItemActionConsume)
}

func (u *Use) release(function, sent string, action int32) error {
	slot := u.self.HotBarSlot()
	u.logger.Info("act", "func", function, "packet", sent)
	return u.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.ReleaseItemTransactionData{
		ActionType:   action,
		HotBarSlot:   int32(slot),
		HeldItem:     u.contents.ItemIn(protocol.WindowIDInventory, slot),
		HeadPosition: u.self.Position(),
	}})
}

// StartSpin and StopSpin are the riptide of a trident: flags in the next tick.
func (u *Use) StartSpin() {
	u.self.AddTickFlag(packet.InputFlagStartSpinAttack)
	u.logger.Info("act", "func", "Use.StartSpin", "packet", "player_auth_input/StartSpinAttack")
}

func (u *Use) StopSpin() {
	u.self.AddTickFlag(packet.InputFlagStopSpinAttack)
	u.logger.Info("act", "func", "Use.StopSpin", "packet", "player_auth_input/StopSpinAttack")
}
