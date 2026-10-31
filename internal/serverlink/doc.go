// Package serverlink is the first layer of the library: it sends packets
// to the server and tells about the packets that arrive. Nothing else.
//
// # Sending
//
// One call sends one packet, as it is. The caller fills in every field.
// The call does not wait for anything: the packet goes into
// gophertunnel's buffer and is flushed within one tick (50 ms), and the
// server's answer, if any, arrives later as a packet of its own.
//
// # Telling
//
// One packet that arrives is told once, to the listener of the system it
// belongs to, found by its type. The packet is passed as it is. A packet
// nobody listens to is dropped. Which packet a packet is an answer to is
// not known here: an answer is told to the system of its own type, which
// can be another system than the one that sent (a click on a chest is
// sent by the operation system; the container_open that follows is told
// to the inventory system).
//
// # What this package does not do
//
// It remembers nothing, turns no numbers into names, passes no data from
// one system to another, answers no packet by itself, and never sends two
// packets as one action. All of that belongs to the next layer.
//
// The one exception is entering the world (package join): the login, the
// resource packs, start_game and the spawn are a fixed sequence of
// packets, so Join does all of it before the first packet is told. What
// gophertunnel and join took from those packets is given by GameData and
// SpawnPosition.
//
// # The 12 systems
//
// Every packet of ordinary play belongs to one of 12 systems
// (packet_systems_v2.drawio). Each system has one file here, holding its
// listeners (the packets that arrive) and its Send functions (the packets
// that are sent). A packet that goes both ways has both.
//
//	connection    entering the world, staying in it, leaving it        connection_system.go
//	dictionary    the lists that turn numbers into names               dictionary_system.go
//	terrain       the blocks around, named by block position           terrain_system.go
//	block entity  blocks that hold data: signs, lecterns, chest lids   block_entity_system.go
//	entity        entities (the bot included), by entity runtime ID    entity_system.go
//	player        what only ever concerns the player                   player_system.go
//	operation     the player's input and its direct replies            operation_system.go
//	inventory     screens with a window ID and the items in them       inventory_system.go
//	text          chat and books                                        text_system.go
//	display       what is drawn over the screen                        display_system.go
//	environment   time, rules, events and sounds of the world          environment_system.go
//	map           maps and the compass                                 map_system.go
//
// A packet of none of the 12 systems (commands, forms, cameras and the
// like) is told to Link.OtherPacket.
//
// # Files
//
//   - doc.go: this.
//   - serverlink.go: Link and Options; Join, Start, Close, Done, Err,
//     GameData, SpawnPosition.
//   - receive_loop.go: reads every packet that arrives until the
//     connection ends, and hands each one out.
//   - tick_clock.go: calls a function once every game tick. It sends
//     nothing by itself.
//   - hand_out.go: the switch that gives an arrived packet to the listener
//     of its system (generated).
//   - *_system.go: one file per system (generated).
//
// The generated files come from make_serverlink.py, which reads the
// system split (systems2.py) and the packet directions (leaves.py).
//
// # Use
//
//	link, err := serverlink.Join(ctx, serverlink.Options{})
//	if err != nil {
//		return err
//	}
//	defer link.Close()
//
//	link.Inventory.ContainerOpen = func(opened *packet.ContainerOpen) {
//		// the next layer remembers which screen is open
//	}
//	link.Start()
//	link.StartTickClock(func(tickNumber uint64) {
//		// the next layer decides what this tick's input is
//		_ = link.SendPlayerAuthInput(input)
//	})
//	<-link.Done()
//
// Listeners are called one at a time, from the receive loop, in the order
// the packets arrived; a listener that takes long holds up the ones after
// it. The tick clock calls its function from a goroutine of its own. Send
// functions may be called from any goroutine. A panic in a listener or in
// the tick function ends the connection, with the panic as the reason in
// Err, rather than the program.
package serverlink
