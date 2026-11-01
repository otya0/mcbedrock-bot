// Names for the types the Bot gives, so that a program can write them: a
// function that takes a *mcbedrock.Move, a field of type mcbedrock.Block.
// They are the same types as the Bot's, under a name of this package.
package mcbedrock

import (
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
)

// The doors.
type (
	Move      = move.Move
	Dig       = dig.Dig
	Hit       = hit.Hit
	Place     = place.Place
	Use       = use.Use
	Choose    = choose.Choose
	Drop      = drop.Drop
	Say       = say.Say
	Container = container.Container
	Respawn   = respawn.Respawn
)

// The records.
type (
	Self     = self.Self
	Contents = contents.Contents
	Blocks   = blocks.Blocks
	Names    = names.Names
	Raw      = raw.Raw
)

// What the records give.
type (
	Block      = blocks.Block
	Item       = contents.Item
	Window     = contents.Window
	TradeOffer = contents.TradeOffer
	Vehicle    = self.Vehicle
	Breaking   = self.Breaking
	Effect     = self.Effect
)
