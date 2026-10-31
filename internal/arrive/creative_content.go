// The creative inventory arrived (while joining).
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) creativeContent(content *packet.CreativeContent) {
	a.contents.SetCreativeItems(content.Items)
}
