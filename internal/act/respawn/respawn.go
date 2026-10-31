// Package respawn presses the respawn button of the death screen: one respawn
// packet in the state ClientReadyToSpawn. The server answers with where to
// come back, and arrive says the bot is back, as a client does by itself.
package respawn

import (
	"errors"
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Respawn struct {
	link   *serverlink.Link
	self   *self.Self
	logger *slog.Logger
}

func New(link *serverlink.Link, selfRecord *self.Self, logger *slog.Logger) *Respawn {
	return &Respawn{link: link, self: selfRecord, logger: logger}
}

// Now presses the button. It can be pressed once the server has said the bot
// may come back (a respawn packet in the state SearchingForSpawn), which
// arrives a moment after death; pressed earlier, the bot would stay dead.
func (r *Respawn) Now() error {
	if !r.self.Dead() {
		return errors.New("respawn now: the bot is not dead")
	}
	if !r.self.CanRespawn() {
		return errors.New("respawn now: the server has not said the bot may come back yet")
	}
	r.self.SetCanRespawn(false)
	r.logger.Info("act", "func", "Respawn.Now", "packet", "respawn/ClientReadyToSpawn")
	return r.link.SendRespawn(&packet.Respawn{State: packet.RespawnStateClientReadyToSpawn, EntityRuntimeID: r.self.EntityRuntimeID()})
}
