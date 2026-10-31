// Package join gets the bot into a world, up to the moment it can act:
// move, use items and so on. Everything needed for that lives in this
// package, and nothing else does.
//
// # Files
//
// The package is split into three files:
//
//   - join.go: the entry point. Open calls the steps below in order and
//     returns HandoverToPlay, what package play needs to start.
//   - dial_lan.go: finds a world on the LAN and connects to it. There is one
//     dial_*.go file per way of connecting. Other ways (by address, Realms,
//     a friend's world) will be added next to it as dial_address.go,
//     dial_realms.go and dial_friend.go.
//   - sign_in.go: signs in to a Microsoft account. It is a separate file
//     because every way of connecting needs it, not only the LAN.
//
// # Steps
//
// Open runs these steps in order:
//
//  1. Find the world on the LAN (dial_lan.go: findWorld).
//  2. Sign in, only if the world asks for it (sign_in.go: signIn).
//  3. Connect to the world (dial_lan.go: dialLAN, using gophertunnel).
//  4. Wait for the spawn (gophertunnel: DoSpawnContext).
//  5. Tell the server the loading screen is gone (join.go:
//     ServerBoundLoadingScreen). A real client sends this; gophertunnel
//     does not.
//  6. Receive packets until the server sends the real spawn position
//     (join.go: receiveSpawnPosition). It comes in a Respawn packet in the
//     state ReadyToSpawn. Minecraft uses this packet both for the first join
//     and for coming back after death, so it arrives even though the bot has
//     never died. The position in StartGame is only a placeholder.
//     Other packets that arrive meanwhile (the inventory, for example) are
//     not dropped: they are handed over to package play in HandoverToPlay.
//  7. Tell the server the bot has spawned (join.go: PlayerAction). This is a
//     PlayerAction packet with the action Respawn, again even on the first
//     join.
//
// Without 5 and 7 the server treats the bot as still loading and ignores
// its movement. After 7 the bot can act.
package join
