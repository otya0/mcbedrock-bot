// Package dig breaks blocks with the left button. Each call puts a block action
// into the next tick's player_auth_input (move sends it); while the button is
// held, move adds the crack of the block every tick.
package dig

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Dig struct {
	self   *self.Self
	logger *slog.Logger
}

func New(selfRecord *self.Self, logger *slog.Logger) *Dig {
	return &Dig{self: selfRecord, logger: logger}
}

// Start holds the left button on a face of a block. Holding it on another block
// first lets go of the one before.
func (d *Dig) Start(position protocol.BlockPos, face int32) error {
	if !blocks.IsFace(face) {
		return fmt.Errorf("dig start: face %d is not 0 to 5", face)
	}
	if breaking, holding := d.self.Breaking(); holding {
		d.self.AddTickBlockAction(protocol.PlayerBlockAction{Action: protocol.PlayerActionAbortBreak, BlockPos: breaking.Position, Face: breaking.Face})
	}
	d.self.AddTickBlockAction(protocol.PlayerBlockAction{Action: protocol.PlayerActionStartBreak, BlockPos: position, Face: face})
	d.self.SetBreaking(self.Breaking{Position: position, Face: face}, true)
	d.logger.Info("act", "func", "Dig.Start", "target", blocks.PositionText(position), "face", face, "packet", "player_auth_input/StartBreak")
	return nil
}

// Stop lets go of the button before the block broke.
func (d *Dig) Stop() error {
	breaking, holding := d.self.Breaking()
	if !holding {
		return errors.New("dig stop: the left button is not held on a block")
	}
	d.self.AddTickBlockAction(protocol.PlayerBlockAction{Action: protocol.PlayerActionAbortBreak, BlockPos: breaking.Position, Face: breaking.Face})
	d.self.SetBreaking(self.Breaking{}, false)
	d.logger.Info("act", "func", "Dig.Stop", "target", blocks.PositionText(breaking.Position), "packet", "player_auth_input/AbortBreak")
	return nil
}

// Finish says the block broke, as a client does when the breaking time is up.
func (d *Dig) Finish(position protocol.BlockPos, face int32) error {
	if !blocks.IsFace(face) {
		return fmt.Errorf("dig finish: face %d is not 0 to 5", face)
	}
	d.self.AddTickBlockAction(protocol.PlayerBlockAction{Action: protocol.PlayerActionPredictDestroyBlock, BlockPos: position, Face: face})
	d.self.AddTickBlockAction(protocol.PlayerBlockAction{Action: protocol.PlayerActionStopBreak, BlockPos: position, Face: face})
	d.self.SetBreaking(self.Breaking{}, false)
	d.logger.Info("act", "func", "Dig.Finish", "target", blocks.PositionText(position), "face", face, "packet", "player_auth_input/PredictDestroyBlock,StopBreak")
	return nil
}

// Continue moves the held button on to the next block.
func (d *Dig) Continue(position protocol.BlockPos, face int32) error {
	if !blocks.IsFace(face) {
		return fmt.Errorf("dig continue: face %d is not 0 to 5", face)
	}
	if _, holding := d.self.Breaking(); !holding {
		return errors.New("dig continue: the left button is not held on a block")
	}
	d.self.AddTickBlockAction(protocol.PlayerBlockAction{Action: protocol.PlayerActionContinueDestroyBlock, BlockPos: position, Face: face})
	d.self.SetBreaking(self.Breaking{Position: position, Face: face}, true)
	d.logger.Info("act", "func", "Dig.Continue", "target", blocks.PositionText(position), "face", face, "packet", "player_auth_input/ContinueDestroyBlock")
	return nil
}
