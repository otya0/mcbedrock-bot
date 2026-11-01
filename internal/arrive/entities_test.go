// Checks that entities appear, move and go in the record, and that the
// vehicle's runtime ID comes from it.
package arrive

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/otya0/mcbedrock-bot/internal/record/entities"
	"github.com/otya0/mcbedrock-bot/internal/record/names"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func arriveForEntities() *Arrive {
	return &Arrive{self: self.New(botRuntimeID, botRuntimeID, [3]float32{}, 0), entities: entities.New(), names: &names.Names{}}
}

func TestEntitiesAppearMoveAndGo(t *testing.T) {
	a := arriveForEntities()
	a.addActor(&packet.AddActor{EntityUniqueID: -5, EntityRuntimeID: 5, EntityType: "minecraft:sheep", Position: mgl32.Vec3{1, 2, 3}})
	a.addPlayer(&packet.AddPlayer{Username: "Alex", EntityRuntimeID: 6, AbilityData: protocol.AbilityData{EntityUniqueID: -6}})
	a.addItemActor(&packet.AddItemActor{EntityUniqueID: -8, EntityRuntimeID: 8})

	// Only z changes: x and y stay.
	a.moveActorDelta(&packet.MoveActorDelta{EntityRuntimeID: 5, PositionZ: protocol.Option[float32](9)})
	if sheep, _ := a.entities.Get(5); sheep.Position != [3]float32{1, 2, 9} {
		t.Errorf("the sheep is at %v, want 1 2 9", sheep.Position)
	}
	a.moveActorAbsolute(&packet.MoveActorAbsolute{EntityRuntimeID: 5, Position: mgl32.Vec3{4, 5, 6}})
	if sheep, _ := a.entities.Get(5); sheep.Position != [3]float32{4, 5, 6} {
		t.Errorf("the sheep is at %v, want 4 5 6", sheep.Position)
	}
	// Another player's move_player moves that player, not the bot.
	a.movePlayer(&packet.MovePlayer{EntityRuntimeID: 6, Position: mgl32.Vec3{7, 8, 9}})
	if alex, _ := a.entities.Get(6); alex.Name != "Alex" || alex.Type != "minecraft:player" || alex.Position != [3]float32{7, 8, 9} {
		t.Errorf("got %+v, want Alex the player at 7 8 9", alex)
	}
	if a.self.Position() != [3]float32{} {
		t.Errorf("the bot moved to %v with another player", a.self.Position())
	}
	if item, _ := a.entities.Get(8); item.Type != "minecraft:item" {
		t.Errorf("got %+v, want an item", item)
	}

	a.removeActor(&packet.RemoveActor{EntityUniqueID: -5})
	if _, ok := a.entities.Get(5); ok {
		t.Error("the sheep is still there after remove_actor")
	}
}

func TestVehicleRuntimeIDFromEntities(t *testing.T) {
	a := arriveForEntities()
	a.addActor(&packet.AddActor{EntityUniqueID: -200, EntityRuntimeID: 20, EntityType: "minecraft:boat"})
	a.setActorLink(&packet.SetActorLink{EntityLink: protocol.EntityLink{RiddenEntityUniqueID: -200, RiderEntityUniqueID: botRuntimeID, Type: protocol.EntityLinkRider}})
	vehicle, riding := a.self.Vehicle()
	if !riding || vehicle.EntityRuntimeID != 20 || vehicle.EntityUniqueID != -200 {
		t.Fatalf("riding %v %+v, want the boat 20 / -200", riding, vehicle)
	}
}
