package serverlink

// receiveLoop reads every packet that arrives and delivers each one (to its
// listener, then to the observers), until the connection ends. Start runs it in a goroutine of its own.
//
// Packets must always be read, wanted or not, or they pile up inside
// gophertunnel. ReadPacket returns an error when the connection has
// ended (closed by Close, dropped, or a disconnect packet from the
// server), and that error becomes Err.
func (link *Link) receiveLoop() {
	defer link.closeOnPanic()
	for {
		arrived, err := link.connection.ReadPacket()
		if err != nil {
			link.finish(err)
			return
		}
		link.deliver(arrived)
	}
}
