// The server corrected where the bot is: write it to self.
package arrive

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (a *Arrive) correctPlayerMovePrediction(correction *packet.CorrectPlayerMovePrediction) {
	// A correction of the vehicle is about the vehicle, not the bot.
	if correction.PredictionType == packet.PredictionTypeVehicle {
		return
	}
	a.self.SetPosition(correction.Position)
}
