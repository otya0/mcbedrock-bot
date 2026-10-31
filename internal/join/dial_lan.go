package join

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-nethernet/discovery"
	"github.com/sandertv/gophertunnel/minecraft"
)

// ErrWorldNotFound means no joinable world was found on the LAN.
var ErrWorldNotFound = errors.New("no world found on the LAN: open a world in Minecraft " +
	"and turn on \"Multiplayer Game\" and \"Visible to LAN Players\" in its settings")

// lanSearchInterval is how often the LAN search checks for replies.
const lanSearchInterval = time.Second

// dialLAN finds a world opened on the same LAN and joins it.
//
// It broadcasts "is anyone hosting a world?" over UDP and connects with
// NetherNet to a world that replies. If worldName is empty, the first world
// found is used; otherwise only a world with that name is chosen. If ctx is
// done before a world is found, dialLAN returns ErrWorldNotFound.
//
// The returned connection has finished logging in but has not spawned yet.
func dialLAN(ctx context.Context, cfg Config) (*minecraft.Conn, error) {
	// The game itself may be using port 7551, so listen on any free port.
	// The search broadcast goes to port 7551 by default.
	finder, err := discovery.ListenConfig{}.Listen("")
	if err != nil {
		return nil, fmt.Errorf("start LAN search: %w", err)
	}
	// finder is also the signaling channel for the connection, so it stays
	// open while connected. Close it here only if connecting fails.
	ok := false
	defer func() {
		if !ok {
			_ = finder.Close()
		}
	}()

	id, world, err := findWorld(ctx, finder, cfg.WorldName)
	if err != nil {
		return nil, err
	}
	cfg.Log.Info("found world", "name", world.LevelName, "host", world.ServerName,
		"players", world.PlayerCount, "max", world.MaxPlayerCount, "version", world.Version)

	d := minecraft.Dialer{ErrorLog: cfg.Log}
	// The host makes a new nonce each time. It must be copied into the Login
	// packet as is, or the host rejects us.
	d.ClientData.Nonce = world.Nonce

	// A world that accepts LAN self-signed auth does not need a Microsoft login.
	if !world.AcceptsSelfSignedAuth {
		src, err := signIn(cfg.TokenPath, cfg.Prompt)
		if err != nil {
			return nil, err
		}
		d.TokenSource = src
	}

	// On the LAN, the network ID is the address to connect to.
	//
	// AllowIdentitylessServer: a LAN world sends no identity in its reply.
	// With the default setting we would reject it with "identityless answer
	// SDP not allowed". We found and chose this LAN world ourselves, so we
	// accept it.
	transport := minecraft.NetherNet{
		Signaling: finder,
		Dialer:    nethernet.Dialer{AllowIdentitylessServer: true},
	}
	conn, err := d.DialContextNetwork(ctx, transport, strconv.FormatUint(id, 10))
	if err != nil {
		return nil, fmt.Errorf("connect to world: %w", err)
	}
	ok = true
	// When the connection closes, close the search as well.
	go func() {
		<-conn.Context().Done()
		_ = finder.Close()
	}()
	return conn, nil
}

// findWorld waits for search replies and returns the first world that matches.
func findWorld(ctx context.Context, finder *discovery.Listener, worldName string) (uint64, *discovery.ServerData, error) {
	tick := time.NewTicker(lanSearchInterval)
	defer tick.Stop()
	for {
		for id, raw := range finder.Responses() {
			world := new(discovery.ServerData)
			if err := world.UnmarshalBinary(raw); err != nil {
				continue // Skip replies we cannot read. Another world may still reply.
			}
			if worldName != "" && world.LevelName != worldName {
				continue
			}
			return id, world, nil
		}
		select {
		case <-tick.C:
		case <-ctx.Done():
			return 0, nil, ErrWorldNotFound
		}
	}
}
