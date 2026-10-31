// The server moved the bot itself (a teleport or a reset): write where it is,
// and say in the next tick's input that it was handled, as a client does.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) movePlayer(moved *packet.MovePlayer) {
	if moved.EntityRuntimeID != a.self.EntityRuntimeID() {
		return
	}
	if moved.Mode != packet.MoveModeTeleport && moved.Mode != packet.MoveModeReset {
		return
	}
	a.self.SetPosition(moved.Position)
	a.self.AddTickFlag(packet.InputFlagHandledTeleport)
}
