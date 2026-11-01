package serverlink

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/otya0/mcbedrock-bot/internal/join"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// DefaultTokenPath is where the token is saved when Options.TokenPath is
// empty. The file is created in the current working directory.
const DefaultTokenPath = "mcbedrock_token.json"

// ErrClosed is the reason in Err when Close ended the connection.
var ErrClosed = errors.New("serverlink: closed by Close")

// Options configures Join. The zero value joins the first world found on
// the LAN.
type Options struct {
	// Address is the server address to connect to. If empty, Join searches
	// the LAN for a world. Connecting by address is not supported yet.
	Address string

	// WorldName is the name of the LAN world to join. If empty, Join uses
	// the first world it finds.
	WorldName string

	// TokenPath is the file that stores the Microsoft account token. If
	// empty, DefaultTokenPath is used. This file is the key to the account:
	// do not share it or commit it to git.
	TokenPath string

	// Prompt is where the login URL and code are written on the first
	// login. If nil, os.Stdout is used.
	Prompt io.Writer

	// Logger receives the logs of entering the world (package join and
	// gophertunnel). If nil, nothing is logged. This package itself logs
	// nothing.
	Logger *slog.Logger
}

// Link is one connection to a world, from the moment the bot can act until
// the connection ends. Make one with Join.
type Link struct {
	connection           *minecraft.Conn
	spawnPosition        [3]float32
	arrivedWhileEntering []packet.Packet

	// The listeners, one set per system. Set them before Start: the receive
	// loop reads them from then on, so they must not change afterwards.
	Connection  ConnectionListeners
	Dictionary  DictionaryListeners
	Terrain     TerrainListeners
	BlockEntity BlockEntityListeners
	Entity      EntityListeners
	Player      PlayerListeners
	Operation   OperationListeners
	Inventory   InventoryListeners
	Text        TextListeners
	Display     DisplayListeners
	Environment EnvironmentListeners
	Map         MapListeners

	// OtherPacket is told a packet of none of the 12 systems (the ones
	// ordinary play does not use: commands, forms, cameras and the like).
	// Leave it nil to drop them.
	OtherPacket func(packet.Packet)

	// observers are told every packet after its listener (see Observe).
	observersMu sync.Mutex
	observers   []func(packet.Packet)

	startOnce  sync.Once
	closeOnce  sync.Once
	finishOnce sync.Once
	done       chan struct{}

	mu  sync.Mutex
	err error
}

// Join enters a world and returns once the bot can act. The packets that
// arrived meanwhile are kept and told when Start is called.
//
// If ctx is done before that, Join gives up and returns an error. Without
// a deadline on ctx, Join keeps searching until it finds a world. The
// first time, a Microsoft account login is needed: the URL and code are
// written to Options.Prompt.
//
// Join never reconnects on its own. If the connection ends, call Join
// again if you want to.
func Join(ctx context.Context, options Options) (*Link, error) {
	config := join.Config{
		Address:   options.Address,
		WorldName: options.WorldName,
		TokenPath: options.TokenPath,
		Prompt:    options.Prompt,
		Log:       options.Logger,
	}
	if config.TokenPath == "" {
		config.TokenPath = DefaultTokenPath
	}
	if config.Prompt == nil {
		config.Prompt = os.Stdout
	}
	if config.Log == nil {
		config.Log = slog.New(slog.DiscardHandler)
	}
	handover, err := join.Open(ctx, config)
	if err != nil {
		return nil, err
	}
	return &Link{
		connection:           handover.Conn,
		spawnPosition:        handover.SpawnPosition,
		arrivedWhileEntering: handover.PacketsReceivedWhileEntering,
		done:                 make(chan struct{}),
	}, nil
}

// Start tells the packets that arrived while entering, in order, then
// starts the receive loop, which tells every packet from then on: each one
// to its listener, then to the observers. Set the listeners before calling
// it. Calling it again does nothing.
func (link *Link) Start() {
	link.startOnce.Do(func() {
		defer link.closeOnPanic()
		for _, arrived := range link.arrivedWhileEntering {
			link.deliver(arrived)
		}
		link.arrivedWhileEntering = nil
		go link.receiveLoop()
	})
}

// closeOnPanic, deferred where listeners, observers and the tick function
// are called, turns a panic in one of them into the end of the connection,
// with the panic as the reason in Err, instead of the end of the whole program.
func (link *Link) closeOnPanic() {
	recovered := recover()
	if recovered == nil {
		return
	}
	link.finish(fmt.Errorf("serverlink: panic in a listener, an observer or the tick function: %v", recovered))
	if link.connection != nil {
		_ = link.connection.Close()
	}
}

// Close ends the connection. It is safe to call more than once, and
// before Start.
func (link *Link) Close() error {
	var err error
	link.closeOnce.Do(func() {
		// ErrClosed first: once the connection closes, the receive loop's
		// ReadPacket fails too, and only the first reason is kept.
		link.finish(ErrClosed)
		err = link.connection.Close()
	})
	return err
}

// Done returns a channel that is closed when the connection has ended,
// whoever ended it.
func (link *Link) Done() <-chan struct{} {
	return link.done
}

// Err returns why the connection ended, or nil while it is still up. A
// disconnect by the server is here too (a disconnect packet is never
// told: gophertunnel closes the connection on it).
func (link *Link) Err() error {
	link.mu.Lock()
	defer link.mu.Unlock()
	return link.err
}

// GameData gives what gophertunnel took from the packets it handled while
// entering: start_game (the bot's entity runtime ID, the game mode, the
// dimension and the rest of the world settings), item_registry (Items),
// chunk_radius_updated (ChunkRadius) and dimension_data (Dimensions).
// Those packets are not told, because they never pass the listeners.
func (link *Link) GameData() minecraft.GameData {
	return link.connection.GameData()
}

// IdentityData gives who the bot is, as the login said: the display name,
// the XUID (empty on a LAN world) and the UUID. gophertunnel took it while
// entering.
func (link *Link) IdentityData() login.IdentityData {
	return link.connection.IdentityData()
}

// SpawnPosition gives the eye position the server sent in the first
// respawn packet while entering (package join, step 6). That packet is
// not told, because join used it. The position in GameData is only a
// placeholder; this is the real one.
func (link *Link) SpawnPosition() [3]float32 {
	return link.spawnPosition
}

// finish records why the connection ended and closes Done. Only the first
// reason is kept.
func (link *Link) finish(err error) {
	link.finishOnce.Do(func() {
		link.mu.Lock()
		link.err = err
		link.mu.Unlock()
		close(link.done)
	})
}
