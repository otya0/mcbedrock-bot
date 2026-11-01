// Package mcbedrock is a Minecraft Bedrock Edition client for a program that
// plays in place of a human: calls become packets to the server, and what
// the server sends can be read back.
//
// A door is one kind of input a player has, as a Bot method: Move (the
// keys), Dig (the left button on a block), Hit (the left button on an
// entity), Place and Use (the right button), Choose (the hotbar), Drop (Q),
// Say (T), Container (the windows), Respawn (the death screen). A door's
// functions are what that input does; the server decides what happens.
//
// What the server tells is read from the records: Self (the bot itself),
// Contents (items and windows), Blocks (the world near the bot, by name),
// Names (the name of any block, item or biome number), Raw (every packet as
// it arrived). Block and biome names need Mojang's data: see download-mojang-data.
//
// A program can also be told every packet that arrives (OnPacket) and every
// tick (OnTick), so that it need not poll the records.
//
// Every type the Bot gives has a name in this package (Move, Self, Block,
// Item, Effect...), so that a program can write it.
package mcbedrock
