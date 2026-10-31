// Package move sends player_auth_input every tick, as a client does: the keys
// that are held, where the bot looks, and what the other doors put in for that
// tick only (block actions, a missed swing).
package move

import (
	"errors"
	"log/slog"
	"math"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// keys is what is held down right now.
type keys struct {
	forward, back, left, right bool
	jump, sprint, sneak        bool
	ascend, descend            bool
	paddleLeft, paddleRight    bool
}

type Move struct {
	link     *serverlink.Link
	self     *self.Self
	contents *contents.Contents
	logger   *slog.Logger

	// The keys and the looking direction, read every tick; only this door needs
	// them. sent is what the last tick sent, to see what started or stopped.
	mu    sync.Mutex
	held  keys
	sent  keys
	yaw   float32
	pitch float32
}

func New(link *serverlink.Link, selfRecord *self.Self, contentsRecord *contents.Contents, logger *slog.Logger, yaw, pitch float32) *Move {
	return &Move{link: link, self: selfRecord, contents: contentsRecord, logger: logger, yaw: yaw, pitch: pitch}
}

func (m *Move) StartForward() { m.hold("Move.StartForward", "Up", &m.held.forward, true) }
func (m *Move) StopForward()  { m.hold("Move.StopForward", "Up", &m.held.forward, false) }
func (m *Move) StartBack()    { m.hold("Move.StartBack", "Down", &m.held.back, true) }
func (m *Move) StopBack()     { m.hold("Move.StopBack", "Down", &m.held.back, false) }
func (m *Move) StartLeft()    { m.hold("Move.StartLeft", "Left", &m.held.left, true) }
func (m *Move) StopLeft()     { m.hold("Move.StopLeft", "Left", &m.held.left, false) }
func (m *Move) StartRight()   { m.hold("Move.StartRight", "Right", &m.held.right, true) }
func (m *Move) StopRight()    { m.hold("Move.StopRight", "Right", &m.held.right, false) }

// Jump, sprint and sneak also say, in the tick they change, that they
// started or stopped (Tick sees the change).
func (m *Move) StartJump() { m.hold("Move.StartJump", "Jumping,StartJumping", &m.held.jump, true) }
func (m *Move) StopJump()  { m.hold("Move.StopJump", "Jumping", &m.held.jump, false) }

// StartSprint holds Ctrl; hold forward too, as a player does.
func (m *Move) StartSprint() {
	m.hold("Move.StartSprint", "Sprinting,StartSprinting", &m.held.sprint, true)
}
func (m *Move) StopSprint() { m.hold("Move.StopSprint", "StopSprinting", &m.held.sprint, false) }
func (m *Move) StartSneak() {
	m.hold("Move.StartSneak", "SneakDown,Sneaking,StartSneaking", &m.held.sneak, true)
}
func (m *Move) StopSneak() { m.hold("Move.StopSneak", "StopSneaking", &m.held.sneak, false) }

// Ascend and descend are the flying keys (space and shift while flying).
func (m *Move) StartAscend()  { m.hold("Move.StartAscend", "Ascend", &m.held.ascend, true) }
func (m *Move) StopAscend()   { m.hold("Move.StopAscend", "Ascend", &m.held.ascend, false) }
func (m *Move) StartDescend() { m.hold("Move.StartDescend", "Descend", &m.held.descend, true) }
func (m *Move) StopDescend()  { m.hold("Move.StopDescend", "Descend", &m.held.descend, false) }

// Swimming, crawling, flying and gliding are said once; the server keeps them.
func (m *Move) StartSwim()  { m.tap("Move.StartSwim", "StartSwimming", packet.InputFlagStartSwimming) }
func (m *Move) StopSwim()   { m.tap("Move.StopSwim", "StopSwimming", packet.InputFlagStopSwimming) }
func (m *Move) StartCrawl() { m.tap("Move.StartCrawl", "StartCrawling", packet.InputFlagStartCrawling) }
func (m *Move) StopCrawl()  { m.tap("Move.StopCrawl", "StopCrawling", packet.InputFlagStopCrawling) }
func (m *Move) StartFly()   { m.tap("Move.StartFly", "StartFlying", packet.InputFlagStartFlying) }
func (m *Move) StopFly()    { m.tap("Move.StopFly", "StopFlying", packet.InputFlagStopFlying) }

// StartGlide opens the elytra as a player does, with jump pressed in the air:
// call it while falling. The server opened it with the jump, not without, nor
// on the ground (a live test in a creative world, 2026-10-31).
func (m *Move) StartGlide() {
	m.self.AddTickFlag(packet.InputFlagJumping)
	m.self.AddTickFlag(packet.InputFlagStartJumping)
	m.tap("Move.StartGlide", "Jumping,StartJumping,StartGliding", packet.InputFlagStartGliding)
}

func (m *Move) StopGlide() { m.tap("Move.StopGlide", "StopGliding", packet.InputFlagStopGliding) }

// The paddles are sent only while the bot rides something.
func (m *Move) StartPaddleLeft() {
	m.hold("Move.StartPaddleLeft", "PaddlingLeft", &m.held.paddleLeft, true)
}
func (m *Move) StopPaddleLeft() {
	m.hold("Move.StopPaddleLeft", "PaddlingLeft", &m.held.paddleLeft, false)
}
func (m *Move) StartPaddleRight() {
	m.hold("Move.StartPaddleRight", "PaddlingRight", &m.held.paddleRight, true)
}
func (m *Move) StopPaddleRight() {
	m.hold("Move.StopPaddleRight", "PaddlingRight", &m.held.paddleRight, false)
}

// Turn looks in a direction, in degrees: yaw 0 south (+Z), 90 west; pitch 90
// straight down, -90 straight up. It stays until the next Turn.
func (m *Move) Turn(yaw, pitch float32) error {
	if math.IsNaN(float64(yaw)) || math.IsNaN(float64(pitch)) || math.IsInf(float64(yaw), 0) || math.IsInf(float64(pitch), 0) {
		return errors.New("turn: the yaw or the pitch is not a number")
	}
	m.mu.Lock()
	m.yaw = yawWithinOneTurn(float64(yaw))
	m.pitch = float32(min(max(float64(pitch), -90), 90))
	yaw, pitch = m.yaw, m.pitch
	m.mu.Unlock()
	m.logger.Info("act", "func", "Move.Turn", "yaw", yaw, "pitch", pitch, "packet", "player_auth_input/Yaw,Pitch")
	return nil
}

func (m *Move) LeaveVehicle() error {
	vehicle, riding := m.self.Vehicle()
	if !riding {
		return errors.New("leave vehicle: the bot rides nothing")
	}
	m.logger.Info("act", "func", "Move.LeaveVehicle", "entity_runtime_id", vehicle.EntityRuntimeID, "packet", "interact/LeaveVehicle")
	return m.link.SendInteract(&packet.Interact{ActionType: packet.InteractActionLeaveVehicle, TargetEntityRuntimeID: vehicle.EntityRuntimeID})
}

// hold presses or lets go of one key.
func (m *Move) hold(function, sent string, key *bool, down bool) {
	m.mu.Lock()
	*key = down
	m.mu.Unlock()
	m.logger.Info("act", "func", function, "packet", "player_auth_input/"+sent)
}

// tap says something in the next tick only.
func (m *Move) tap(function, sent string, flag int) {
	m.self.AddTickFlag(flag)
	m.logger.Info("act", "func", function, "packet", "player_auth_input/"+sent)
}

// Tick sends this tick's player_auth_input. It is not a method so that
// Bot.Move() does not show it: layer 1's tick clock calls it 20 times a
// second. Nothing is sent while the bot is dead (what was added for the tick
// is dropped), and no key moves the bot while a window is open.
func Tick(m *Move, tick uint64) {
	if m.self.Dead() {
		m.self.TakeTickInput()
		return
	}
	_, open := m.contents.OpenWindow()
	m.mu.Lock()
	held, yaw, pitch := m.held, m.yaw, m.pitch
	if open {
		held = keys{}
	}
	before := m.sent
	m.sent = held
	m.mu.Unlock()
	flags, blockActions := m.self.TakeTickInput()
	if breaking, holding := m.self.Breaking(); holding && len(blockActions) == 0 {
		blockActions = append(blockActions, protocol.PlayerBlockAction{Action: protocol.PlayerActionCrackBreak, BlockPos: breaking.Position, Face: breaking.Face})
	}
	vehicle, riding := m.self.Vehicle()

	input := heldFlags(held, riding)
	for _, flag := range append(flags, changeFlags(before, held)...) {
		input.Set(flag)
	}
	authInput := &packet.PlayerAuthInput{
		Pitch:              pitch,
		Yaw:                yaw,
		HeadYaw:            yaw,
		Position:           m.self.Position(),
		MoveVector:         stick(held),
		RawMoveVector:      stick(held),
		AnalogueMoveVector: stick(held),
		InputMode:          packet.InputModeMouse,
		PlayMode:           packet.PlayModeNormal,
		InteractionModel:   packet.InteractionModelCrosshair,
		InteractPitch:      pitch,
		InteractYaw:        yaw,
		Tick:               tick,
		CameraOrientation:  lookDirection(yaw, pitch),
	}
	if len(blockActions) > 0 {
		input.Set(packet.InputFlagPerformBlockActions)
		authInput.BlockActions = protocol.Option(blockActions)
	}
	if riding {
		input.Set(packet.InputFlagClientPredictedVehicle)
		authInput.ClientPredictedVehicle = protocol.Option(vehicle.EntityUniqueID)
		authInput.VehicleRotation = protocol.Option(mgl32.Vec2{0, yaw})
	}
	authInput.InputData = input
	_ = m.link.SendPlayerAuthInput(authInput)
}

// heldFlags turns the held keys into input flags. A real client always sets
// BlockBreakingDelayEnabled; the paddles count only in a vehicle.
func heldFlags(held keys, riding bool) protocol.InputFlags {
	input := protocol.NewInputFlags(packet.InputFlagCount)
	input.Set(packet.InputFlagBlockBreakingDelayEnabled)
	keyFlags := []struct {
		down bool
		flag int
	}{
		{held.forward, packet.InputFlagUp},
		{held.back, packet.InputFlagDown},
		{held.left, packet.InputFlagLeft},
		{held.right, packet.InputFlagRight},
		{held.jump, packet.InputFlagJumping},
		{held.sprint, packet.InputFlagSprinting},
		{held.sneak, packet.InputFlagSneakDown},
		{held.sneak, packet.InputFlagSneaking},
		{held.ascend, packet.InputFlagAscend},
		{held.descend, packet.InputFlagDescend},
		{riding && held.paddleLeft, packet.InputFlagPaddlingLeft},
		{riding && held.paddleRight, packet.InputFlagPaddlingRight},
	}
	for _, key := range keyFlags {
		if key.down {
			input.Set(key.flag)
		}
	}
	return input
}

// changeFlags says what started or stopped since the last tick, as a client
// does the tick a key goes down or up.
func changeFlags(before, now keys) []int {
	var flags []int
	if now.jump && !before.jump {
		flags = append(flags, packet.InputFlagStartJumping)
	}
	if now.sprint && !before.sprint {
		flags = append(flags, packet.InputFlagStartSprinting)
	}
	if !now.sprint && before.sprint {
		flags = append(flags, packet.InputFlagStopSprinting)
	}
	if now.sneak && !before.sneak {
		flags = append(flags, packet.InputFlagStartSneaking)
	}
	if !now.sneak && before.sneak {
		flags = append(flags, packet.InputFlagStopSneaking)
	}
	return flags
}

// stick is the movement as a joystick: forward is +1 on the second axis,
// left is +1 on the first, as a client sends it.
func stick(held keys) [2]float32 {
	var direction [2]float32
	if held.forward {
		direction[1]++
	}
	if held.back {
		direction[1]--
	}
	if held.left {
		direction[0]++
	}
	if held.right {
		direction[0]--
	}
	return direction
}

// lookDirection is the unit vector of a looking direction.
func lookDirection(yaw, pitch float32) [3]float32 {
	yawRadians := float64(yaw) * math.Pi / 180
	pitchRadians := float64(pitch) * math.Pi / 180
	return [3]float32{
		float32(-math.Sin(yawRadians) * math.Cos(pitchRadians)),
		float32(-math.Sin(pitchRadians)),
		float32(math.Cos(yawRadians) * math.Cos(pitchRadians)),
	}
}

// yawWithinOneTurn gives the same direction from -180 up to (not including) 180.
func yawWithinOneTurn(yaw float64) float32 {
	yaw = math.Mod(yaw, 360)
	if yaw >= 180 {
		yaw -= 360
	} else if yaw < -180 {
		yaw += 360
	}
	// Rounding to float32 can turn 179.99999 into 180, which is -180.
	if turned := float32(yaw); turned < 180 {
		return turned
	}
	return -180
}
