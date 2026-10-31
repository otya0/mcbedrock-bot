// The item list arrived (while joining): it turns item numbers into names.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) itemRegistry(registry *packet.ItemRegistry) {
	a.contents.SetItemNames(registry.Items)
	a.names.SetItems(registry.Items)
}
