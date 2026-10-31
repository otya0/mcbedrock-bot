// Package self keeps what the server last told about the bot itself, and
// what the other doors want to ride in the next tick's player_auth_input.
package self

import (
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Vehicle is what the bot rides: player_auth_input names it by unique ID,
// interact by runtime ID.
type Vehicle struct {
	EntityUniqueID  int64
	EntityRuntimeID uint64
}

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
	return &Self{entityRuntimeID: entityRuntimeID, entityUniqueID: entityUniqueID, position: position, dimension: dimension}
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
