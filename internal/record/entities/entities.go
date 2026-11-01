// Package entities keeps the entities around the bot that the server has told
// about (animals, monsters, other players, items on the ground), by runtime
// ID, from when they appear until the server removes them.
package entities

import (
	"cmp"
	"slices"
	"sync"
)

// Entity is one entity. RuntimeID is the number the doors take (Hit.Entity,
// Use.OnEntity, Container.OpenOnEntity); UniqueID the number of links and
// removals. Type is its kind ("minecraft:zombie"; "minecraft:player" for a
// player, "minecraft:item" for an item on the ground). Name is a player's
// name, or the item's name for an item on the ground. Position is where the
// server last said it is; for a player, as for the bot, that is the eyes.
type Entity struct {
	RuntimeID uint64
	UniqueID  int64
	Type      string
	Name      string
	Position  [3]float32
}

// MaxEntities is far more than a world shows at once. Past it, new entities
// are refused (and counted by Refused); the ones kept are never dropped.
const MaxEntities = 1 << 16

type Entities struct {
	mu       sync.Mutex
	byID     map[uint64]Entity
	byUnique map[int64]uint64
	refused  int
}

func New() *Entities {
	return &Entities{byID: map[uint64]Entity{}, byUnique: map[int64]uint64{}}
}

// Get gives the entity with this runtime ID, if the bot knows it.
func (e *Entities) Get(runtimeID uint64) (Entity, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entity, ok := e.byID[runtimeID]
	return entity, ok
}

// ByUniqueID gives the entity with this unique ID, if the bot knows it.
func (e *Entities) ByUniqueID(uniqueID int64) (Entity, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	runtimeID, ok := e.byUnique[uniqueID]
	if !ok {
		return Entity{}, false
	}
	return e.byID[runtimeID], true
}

// All gives every entity the bot knows, by runtime ID from low to high.
func (e *Entities) All() []Entity {
	e.mu.Lock()
	defer e.mu.Unlock()
	all := make([]Entity, 0, len(e.byID))
	for _, entity := range e.byID {
		all = append(all, entity)
	}
	slices.SortFunc(all, func(a, b Entity) int { return cmp.Compare(a.RuntimeID, b.RuntimeID) })
	return all
}

// Refused is how many new entities were not kept because MaxEntities were
// already kept. It stays 0 unless a server sends far more than a world has.
func (e *Entities) Refused() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.refused
}

// Add keeps an entity that appeared; one with the same runtime ID is replaced.
func (e *Entities) Add(entity Entity) {
	e.mu.Lock()
	defer e.mu.Unlock()
	old, known := e.byID[entity.RuntimeID]
	if !known && len(e.byID) >= MaxEntities {
		e.refused++
		return
	}
	if known {
		delete(e.byUnique, old.UniqueID)
	}
	e.byID[entity.RuntimeID] = entity
	// A unique ID of 0 means the server left it out: it names nothing.
	if entity.UniqueID != 0 {
		e.byUnique[entity.UniqueID] = entity.RuntimeID
	}
}

// RemoveByUniqueID forgets an entity that disappeared (remove_actor names it
// by unique ID).
func (e *Entities) RemoveByUniqueID(uniqueID int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	runtimeID, ok := e.byUnique[uniqueID]
	if !ok {
		return
	}
	delete(e.byUnique, uniqueID)
	delete(e.byID, runtimeID)
}

// SetPosition moves a known entity; an unknown one is left unknown.
func (e *Entities) SetPosition(runtimeID uint64, position [3]float32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entity, ok := e.byID[runtimeID]
	if !ok {
		return
	}
	entity.Position = position
	e.byID[runtimeID] = entity
}

// Forget drops every entity: the dimension changed.
func (e *Entities) Forget() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.byID = map[uint64]Entity{}
	e.byUnique = map[int64]uint64{}
}
