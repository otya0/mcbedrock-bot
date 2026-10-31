// The options of the enchanting table arrived.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) playerEnchantOptions(options *packet.PlayerEnchantOptions) {
	a.contents.SetEnchantOptions(options.Options)
}
