# mcbedrock-bot

mcbedrock-bot is a Go library for **Minecraft Bedrock Edition bots**: it gives
a program the **basic actions** of a player, so that the program can play in
place of a human. Each call is what a player does with the keys, the mouse
buttons, the hotbar or a window: walk, jump, dig, place a block, use an item,
chat, open a chest, respawn. What the server sends back can be read: where the
bot is, its items, the blocks and biomes around it.

mcbedrock-bot does not decide anything. It has no path finding, no fighting
and no plans: those belong to the program built on top of it.

> NOT AN OFFICIAL MINECRAFT PRODUCT. NOT APPROVED BY OR ASSOCIATED WITH MOJANG OR MICROSOFT.

## Rules

- Join only worlds you own, or worlds whose owner has said yes.
- mcbedrock-bot sends what a player's client sends, and nothing on its own: no cheats.
- Follow the [Minecraft EULA](https://www.minecraft.net/eula) and the rules of the world you join.

## Requirements

- Go 1.26 or later
- A Microsoft account that owns Minecraft Bedrock Edition
- A world opened to LAN ("Visible to LAN Players") on the same network

mcbedrock-bot is tested on Windows with Minecraft 1.26.60.

## Install

```
go get github.com/otya0/mcbedrock-bot
```

Then take Mojang's data once (see below):

```
go run github.com/otya0/mcbedrock-bot/cmd/download-mojang-data@latest
```

## Example

The package name is `mcbedrock`.

```go
import "github.com/otya0/mcbedrock-bot"

bot, err := mcbedrock.Join(ctx, mcbedrock.Options{})
if err != nil {
	log.Fatal(err)
}
defer bot.Close()

bot.Move().StartForward()
time.Sleep(2 * time.Second)
bot.Move().StopForward()

block, ok := bot.Blocks().At(protocol.BlockPos{10, 64, -3})
if ok {
	fmt.Println(block.Name, block.Biome) // minecraft:stone minecraft:plains
}
```

## Mojang's data

The server sends blocks and biomes as numbers. To give them names,
mcbedrock-bot reads Mojang's own lists from
[Mojang/bedrock-samples](https://github.com/Mojang/bedrock-samples).
These files belong to Mojang, so mcbedrock-bot does not include them:
`download-mojang-data` takes them from Mojang's repository to your computer
(the user cache folder, `mcbedrock-bot/mojang`). `Join` stops with an error
until they are there.

## Account

The first `Join` prints a Microsoft login URL and a code. After you sign in,
the token is saved in `mcbedrock_token.json`. **This file is the key to your
account: never share it or commit it.**

## Logs

mcbedrock-bot prints nothing. Give `Options.Logger` a `*slog.Logger` to see
each call it makes.

## How it is built

![The parts of mcbedrock-bot](docs/system.png)

- **mcbedrock** (`bot.go`, `join.go`) is the only package your program imports.
- **act** sends what a player does. Each door is one kind of input (`Move`,
  `Dig`, `Hit`, `Place`, `Use`, `Choose`, `Drop`, `Say`, `Container`,
  `Respawn`); the server decides what happens.
- **record** remembers what the server has told (`Self`, `Contents`,
  `Blocks`, `Names`, `Raw`) and answers when asked.
- **arrive** reads each packet that arrives, writes the records, and replies
  where a client replies by itself.
- **serverlink** and **join** sign in, connect, send packets and tell which
  ones arrived. They are built on
  [gophertunnel](https://github.com/Sandertv/gophertunnel).

## Questions and bugs

Please open an issue on GitHub.

## License

MIT, see [LICENSE](LICENSE). Mojang's data is not covered: it belongs to Mojang.
