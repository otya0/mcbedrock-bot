package join

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Config configures Open. The caller (Join in the root package) fills it in.
type Config struct {
	// Address is the server address. If empty, Open searches the LAN for a world.
	Address string
	// WorldName is the name of the LAN world to join. If empty, the first world found is used.
	WorldName string
	// TokenPath is the file that stores the Microsoft account token.
	TokenPath string
	// Prompt is where the login URL and code are written on the first login.
	Prompt io.Writer
	// Log receives logs from gophertunnel and this package. It must not be nil.
	Log *slog.Logger
}

// HandoverToPlay is what the bot got from the server while entering the
// world, handed over to package play so that play can start. It holds only
// what play needs.
type HandoverToPlay struct {
	// Conn is the connection to the world.
	Conn *minecraft.Conn
	// SpawnPosition is the real eye position received in Step 6. The
	// position in Conn.GameData() is only a placeholder; use this one.
	SpawnPosition [3]float32
	// PacketsReceivedWhileEntering are the packets that arrived in Step 6
	// while waiting for the spawn position, in the order they arrived. The
	// inventory, for example, arrives here and is sent only once. Package
	// play handles them before any packet received later.
	PacketsReceivedWhileEntering []packet.Packet
}

// Open enters a world and, once the bot can act, returns what package play
// needs.
// The steps are listed in doc.go and marked "Step N" below.
//
// If ctx is done before that, Open gives up and returns an error.
func Open(ctx context.Context, cfg Config) (*HandoverToPlay, error) {
	if cfg.Address != "" {
		return nil, errors.New("connecting by address is not supported yet")
	}

	// Steps 1, 2 and 3: find the world, sign in if needed, and connect.
	conn, err := dialLAN(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Step 4: wait for the spawn.
	if err := conn.DoSpawnContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("spawn: %w", err)
	}

	// Step 5: tell the server the loading screen is gone. A real client sends
	// this once after spawning; gophertunnel does not.
	if err := conn.WritePacket(&packet.ServerBoundLoadingScreen{Type: packet.LoadingScreenTypeEnd}); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("send loading screen end: %w", err)
	}

	// Step 6: receive the real spawn position, and keep the other packets
	// that arrive meanwhile for package play.
	pos, received, err := receiveSpawnPosition(ctx, conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	cfg.Log.Info("spawn position received", "position", pos)

	// Step 7: tell the server the bot has spawned. This is a PlayerAction
	// packet with the action Respawn, even on the first join.
	spawned := &packet.PlayerAction{
		EntityRuntimeID: conn.GameData().EntityRuntimeID,
		ActionType:      protocol.PlayerActionRespawn,
	}
	if err := conn.WritePacket(spawned); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("send spawned: %w", err)
	}

	return &HandoverToPlay{
		Conn:                         conn,
		SpawnPosition:                pos,
		PacketsReceivedWhileEntering: received,
	}, nil
}

// receiveSpawnPosition receives packets until the server sends the real spawn
// position, and returns it. Other packets that arrive meanwhile are not
// dropped: they are returned too, in the order they arrived, so that package
// play can handle them.
//
// The position comes in a Respawn packet in the state ReadyToSpawn. Minecraft
// uses this packet both for the first join and for coming back after death,
// so it arrives here even though the bot has never died.
//
// If ctx is done while waiting, the connection is closed so that ReadPacket
// returns, and ctx's error is returned.
func receiveSpawnPosition(ctx context.Context, conn *minecraft.Conn) ([3]float32, []packet.Packet, error) {
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	var received []packet.Packet
	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			return [3]float32{}, nil, fmt.Errorf("wait for spawn position: %w", err)
		}
		if r, ok := pk.(*packet.Respawn); ok && r.State == packet.RespawnStateReadyToSpawn {
			return r.Position, received, nil
		}
		received = append(received, pk)
	}
}
