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
	"github.com/otya0/mcbedrock-bot/internal/record/names"
	"github.com/otya0/mcbedrock-bot/internal/record/raw"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
)

// WorldName: the LAN world to join (empty: the first one found). TokenPath: the
// file of the account token (empty: mcbedrock_token.json; never share or commit
// it). MojangDataFolder: where download-mojang-data saved Mojang's data (empty:
// the usual cache folder). Logger: where the called functions are logged (nil: nothing).
type Options struct {
	WorldName        string
	TokenPath        string
	MojangDataFolder string
	Logger           *slog.Logger
}

// Join enters the world and returns once the bot can act. The first time, the
// Microsoft login URL and code are printed to the terminal. Without Mojang's
// data it stops before joining, saying how to take it, as a server stops
// until its EULA is agreed to.
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
	link, err := serverlink.Join(ctx, serverlink.Options{WorldName: options.WorldName, TokenPath: options.TokenPath})
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
	arrive.Connect(link, b.raw, b.self, b.contents, b.blocks, b.names)
	link.Start()
	// move.Tick is the tick clock's job, not a user's: it is not a method of Move.
	link.StartTickClock(func(tick uint64) { move.Tick(b.move, tick) })
	return b, nil
}
