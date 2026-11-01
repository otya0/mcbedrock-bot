// Bot is one bot in one world. Its methods are everything a user can do (the
// ten doors) and read (self, contents, blocks, names, raw). Join makes it.
package mcbedrock

import (
	"sync"

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
	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/names"
	"github.com/otya0/mcbedrock-bot/internal/record/raw"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
)

type Bot struct {
	link *serverlink.Link

	raw      *raw.Raw
	self     *self.Self
	contents *contents.Contents
	blocks   *blocks.Blocks
	names    *names.Names

	move      *move.Move
	dig       *dig.Dig
	hit       *hit.Hit
	place     *place.Place
	use       *use.Use
	choose    *choose.Choose
	drop      *drop.Drop
	say       *say.Say
	container *container.Container
	respawn   *respawn.Respawn

	// The OnTick functions, called by the tick clock.
	tickMu        sync.Mutex
	tickObservers []func(tick uint64)
}

// What the bot does: one door per kind of input.
func (b *Bot) Move() *move.Move                { return b.move }
func (b *Bot) Dig() *dig.Dig                   { return b.dig }
func (b *Bot) Hit() *hit.Hit                   { return b.hit }
func (b *Bot) Place() *place.Place             { return b.place }
func (b *Bot) Use() *use.Use                   { return b.use }
func (b *Bot) Choose() *choose.Choose          { return b.choose }
func (b *Bot) Drop() *drop.Drop                { return b.drop }
func (b *Bot) Say() *say.Say                   { return b.say }
func (b *Bot) Container() *container.Container { return b.container }
func (b *Bot) Respawn() *respawn.Respawn       { return b.respawn }

// Output: the bot itself, the items and windows, the blocks with their names,
// the names of any number (for packets in Raw), and every packet as it arrived.
func (b *Bot) Self() *self.Self             { return b.self }
func (b *Bot) Contents() *contents.Contents { return b.contents }
func (b *Bot) Blocks() *blocks.Blocks       { return b.blocks }
func (b *Bot) Names() *names.Names          { return b.names }
func (b *Bot) Raw() *raw.Raw                { return b.raw }

// Close leaves the world.
func (b *Bot) Close() error { return b.link.Close() }

// Done is closed when the connection has ended, and Err says why.
func (b *Bot) Done() <-chan struct{} { return b.link.Done() }
func (b *Bot) Err() error            { return b.link.Err() }
