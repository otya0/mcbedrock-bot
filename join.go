// Join enters a world and makes the Bot: the records, the doors, and arrive
// listening to layer 1. Options is what Join needs.
package mcbedrock

import (
	"context"
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/act/choose"
	"github.com/otya0/mcbedrock-bot/internal/act/container"
	"github.com/otya0/mcbedrock-bot/internal/act/dig"
	"github.com/otya0/mcbedrock-bot/internal/act/drop"
	"github.com/otya0/mcbedrock-bot/internal/act/hit"
	"github.com/otya0/mcbedrock-bot/internal/act/move"
	"github.com/otya0/mcbedrock-bot/internal/act/place"
	"github.com/otya0/mcbedrock-bot/internal/act/respawn"
	"github.com/otya0/mcbedrock-bot/internal/act/say"
	"github.com/otya0/mcbedrock-bot/internal/act/use"
	"github.com/otya0/mcbedrock-bot/internal/arrive"
	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/entities"
	"github.com/otya0/mcbedrock-bot/internal/record/names"
	"github.com/otya0/mcbedrock-bot/internal/record/raw"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// WorldName: the LAN world to join (empty: the first one found). TokenPath: the
// file of the account token (empty: mcbedrock_token.json; never share or commit
// it). MojangDataFolder: where download-mojang-data saved Mojang's data (empty:
// the usual cache folder). Logger: where joining (each step, and gophertunnel's
// own logs) and the called functions are logged (nil: nothing).
// OnPacket: told every packet from the first one, as Bot.OnPacket is; it starts
// while Join is still going, before the Bot is there.
type Options struct {
	WorldName        string
	TokenPath        string
	MojangDataFolder string
	Logger           *slog.Logger
	OnPacket         func(packet.Packet)
}

// Join enters the world and returns once the bot can act. A LAN world that
// accepts self-signed sign-in needs no Microsoft account; one that asks for
// an Xbox Live sign-in gets it, and the first time the login URL and code are
// printed to the terminal. Without Mojang's data Join stops before joining,
// saying how to take it, as a server stops until its EULA is agreed to.
func Join(ctx context.Context, options Options) (*Bot, error) {
	mojangFolder := options.MojangDataFolder
	if mojangFolder == "" {
		folder, err := names.MojangDataFolder()
		if err != nil {
			return nil, err
		}
		mojangFolder = folder
	}
	nameTable, err := names.New(mojangFolder)
	if err != nil {
		return nil, err
	}
	link, err := serverlink.Join(ctx, serverlink.Options{WorldName: options.WorldName, TokenPath: options.TokenPath, Logger: options.Logger})
	if err != nil {
		return nil, err
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	gameData := link.GameData()
	b := &Bot{
		link:     link,
		raw:      raw.New(),
		self:     self.New(gameData.EntityRuntimeID, gameData.EntityUniqueID, link.SpawnPosition(), gameData.Dimension),
		contents: contents.New(gameData.Items),
		blocks:   blocks.New(),
		names:    nameTable,
		entities: entities.New(),
	}
	b.names.SetItems(gameData.Items)
	b.names.SetCustomBlocks(gameData.CustomBlocks)
	b.move = move.New(link, b.self, b.contents, logger, gameData.Yaw, gameData.Pitch)
	b.dig = dig.New(b.self, logger)
	b.hit = hit.New(link, b.self, b.contents, logger)
	b.place = place.New(link, b.self, b.contents, b.blocks, logger)
	b.use = use.New(link, b.self, b.contents, b.blocks, logger)
	b.choose = choose.New(link, b.self, b.contents, logger)
	b.drop = drop.New(link, b.self, b.contents, logger)
	b.say = say.New(link, logger)
	b.container = container.New(link, b.self, b.contents, b.blocks, logger)
	b.respawn = respawn.New(link, b.self, logger)
	arrive.Connect(link, b.raw, b.self, b.contents, b.blocks, b.names, b.entities)
	if options.OnPacket != nil {
		link.Observe(options.OnPacket)
	}
	link.Start()
	link.StartTickClock(b.tick)
	return b, nil
}
