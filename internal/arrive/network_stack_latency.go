// The server measures the latency: send the same timestamp back, as a client
// does. gophertunnel does not answer by itself.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) networkStackLatency(request *packet.NetworkStackLatency) {
	if request.NeedsResponse {
		_ = a.link.SendNetworkStackLatency(&packet.NetworkStackLatency{Timestamp: request.Timestamp})
	}
}
