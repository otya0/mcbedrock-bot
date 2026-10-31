// The window closed: the answer to Close, or the server closing it.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) containerClose(closed *packet.ContainerClose) {
	a.contents.SetClosed(closed.WindowID)
}
