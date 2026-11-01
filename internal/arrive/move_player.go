// A player moved. Another player: write where it is now. The bot itself, moved
// by the server (a teleport or a reset): write where it is, and say in the
// next tick's input that it was handled, as a client does.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) movePlayer(moved *packet.MovePlayer) {
	if moved.EntityRuntimeID != a.self.EntityRuntimeID() {
		a.entities.SetPosition(moved.EntityRuntimeID, moved.Position)
		return
	}
	if moved.Mode != packet.MoveModeTeleport && moved.Mode != packet.MoveModeReset {
		return
	}
	a.self.SetPosition(moved.Position)
	a.self.AddTickFlag(packet.InputFlagHandledTeleport)
}
