// Package names turns the numbers the server uses into the names a human
// reads: blocks ("minecraft:oak_door" and its property values), items and
// biomes. The game's own blocks and biomes come from Mojang's data
// (mojang_data.go); the world's own ones, and every item, from the server.
package names

import (
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Names struct {
	mu     sync.Mutex
	blocks map[uint32]blockState
	items  map[int32]string
	biomes map[int32]string
}

// New reads Mojang's data from a folder (MojangDataFolder gives the usual
// one) and makes the tables of the game's own blocks and biomes. Without the
// data it fails, saying how to take it.
func New(mojangFolder string) (*Names, error) {
	mojang, err := readMojangData(mojangFolder)
	if err != nil {
		return nil, err
	}
	n := &Names{blocks: map[uint32]blockState{}, items: map[int32]string{}, biomes: map[int32]string{}}
	for _, entry := range mojang.blockList {
		n.addBlock(entry.name, entry.properties, mojang.blockProperties)
	}
	for id, name := range mojang.biomeList {
		n.biomes[id] = name
	}
	return n, nil
}

// SetItems takes the item list the server sends (item_registry): the game's
// items and the world's own.
func (n *Names) SetItems(items []protocol.ItemEntry) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, item := range items {
		n.items[int32(item.RuntimeID)] = item.Name
	}
}

// Item is the name of an item number ("minecraft:dirt"), "" if the server did not list it.
func (n *Names) Item(networkID int32) string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.items[networkID]
}
