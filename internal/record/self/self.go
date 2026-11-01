// Package self keeps what the server last told about the bot itself, and
// what the other doors want to ride in the next tick's player_auth_input.
package self

import (
	"cmp"
	"slices"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Vehicle is what the bot rides: player_auth_input names it by unique ID,
// interact by runtime ID.
type Vehicle struct {
	EntityUniqueID  int64
	EntityRuntimeID uint64
}

// Effect is one effect on the bot: ID is packet.EffectSpeed, packet.EffectPoison
// and the others, Amplifier the level minus one, Duration the ticks left when
// the server last told.
type Effect struct {
	ID        int32
	Amplifier int32
	Duration  int32
}

// maxEffects is more than all the effects of the game together: a server that
// sends more distinct ones gets no more kept.
const maxEffects = 64

// Breaking is the block the left button is held on.
type Breaking struct {
	Position protocol.BlockPos
	Face     int32
}

type Self struct {
	mu              sync.Mutex
	entityRuntimeID uint64
	entityUniqueID  int64
	position        [3]float32
	dimension       int32
	hotBarSlot      int
	health          float32
	maxHealth       float32
	food            float32
	effects         map[int32]Effect
	dead            bool
	canRespawn      bool
	vehicle         Vehicle
	riding          bool
	breaking        Breaking
	holdingBreak    bool
	tickFlags       []int
	tickActions     []protocol.PlayerBlockAction
}

// New starts from what the server gave while joining.
func New(entityRuntimeID uint64, entityUniqueID int64, position [3]float32, dimension int32) *Self {
	return &Self{entityRuntimeID: entityRuntimeID, entityUniqueID: entityUniqueID, position: position, dimension: dimension, effects: map[int32]Effect{}}
}

func (s *Self) EntityRuntimeID() uint64 { return s.entityRuntimeID }
func (s *Self) EntityUniqueID() int64   { return s.entityUniqueID }

// Position is where the server last said the bot is (the eyes).
func (s *Self) Position() [3]float32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.position
}

func (s *Self) SetPosition(position [3]float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.position = position
}

// Dimension is 0 for the overworld, 1 for the nether, 2 for the end.
func (s *Self) Dimension() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dimension
}

func (s *Self) SetDimension(dimension int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dimension = dimension
}

// HotBarSlot is the selected hotbar slot, 0 to 8.
func (s *Self) HotBarSlot() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hotBarSlot
}

// SetHotBarSlot keeps only 0 to 8, whatever the server or a caller says.
func (s *Self) SetHotBarSlot(slot int) {
	if slot < 0 || slot > 8 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hotBarSlot = slot
}

// Health is the bot's health now and the most it can be, in half hearts (20 is
// ten hearts). Both are 0 until the server has told them.
func (s *Self) Health() (now, maximum float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.health, s.maxHealth
}

func (s *Self) SetHealth(now, maximum float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.health, s.maxHealth = now, maximum
}

// Food is the hunger bar, 0 to 20. It is 0 until the server has told it.
func (s *Self) Food() float32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.food
}

func (s *Self) SetFood(food float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.food = food
}

// Effects are the effects on the bot now, by ID from low to high.
func (s *Self) Effects() []Effect {
	s.mu.Lock()
	defer s.mu.Unlock()
	effects := make([]Effect, 0, len(s.effects))
	for _, effect := range s.effects {
		effects = append(effects, effect)
	}
	slices.SortFunc(effects, func(a, b Effect) int { return cmp.Compare(a.ID, b.ID) })
	return effects
}

// SetEffect adds an effect or changes the one with the same ID.
func (s *Self) SetEffect(effect Effect) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, known := s.effects[effect.ID]; !known && len(s.effects) >= maxEffects {
		return
	}
	s.effects[effect.ID] = effect
}

func (s *Self) RemoveEffect(id int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.effects, id)
}

// Dead is true from death_info until the bot is back to life.
func (s *Self) Dead() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dead
}

func (s *Self) SetDead(dead bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dead = dead
}

// CanRespawn is true while the bot is dead and the server has said it may
// come back: the respawn button can be pressed.
func (s *Self) CanRespawn() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.canRespawn
}

func (s *Self) SetCanRespawn(can bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.canRespawn = can
}

// Vehicle is what the bot rides, if it rides something.
func (s *Self) Vehicle() (Vehicle, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vehicle, s.riding
}

func (s *Self) SetVehicle(vehicle Vehicle, riding bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vehicle, s.riding = vehicle, riding
}

// Breaking is the block the left button is held on, if it is held.
func (s *Self) Breaking() (Breaking, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.breaking, s.holdingBreak
}

func (s *Self) SetBreaking(breaking Breaking, holding bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.breaking, s.holdingBreak = breaking, holding
}

// AddTickFlag and AddTickBlockAction put something into the next tick's
// input only (a missed swing, a block action).
func (s *Self) AddTickFlag(flag int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickFlags = append(s.tickFlags, flag)
}

func (s *Self) AddTickBlockAction(action protocol.PlayerBlockAction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickActions = append(s.tickActions, action)
}

// TakeTickInput gives what was added for the next tick and forgets it.
func (s *Self) TakeTickInput() ([]int, []protocol.PlayerBlockAction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	flags, actions := s.tickFlags, s.tickActions
	s.tickFlags, s.tickActions = nil, nil
	return flags, actions
}
