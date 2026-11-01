// Checks that entities are kept by runtime ID, removed by unique ID, moved,
// forgotten, and refused (never dropped) past MaxEntities.
package entities

import "testing"

func TestAddMoveRemove(t *testing.T) {
	e := New()
	e.Add(Entity{RuntimeID: 9, UniqueID: -90, Type: "minecraft:zombie"})
	e.Add(Entity{RuntimeID: 3, UniqueID: -30, Type: "minecraft:player", Name: "Steve"})
	if all := e.All(); len(all) != 2 || all[0].RuntimeID != 3 || all[1].RuntimeID != 9 {
		t.Fatalf("All gave %+v, want runtime IDs 3 then 9", all)
	}
	e.SetPosition(9, [3]float32{1, 2, 3})
	if zombie, _ := e.Get(9); zombie.Position != [3]float32{1, 2, 3} {
		t.Errorf("the zombie is at %v, want 1 2 3", zombie.Position)
	}
	if player, ok := e.ByUniqueID(-30); !ok || player.Name != "Steve" {
		t.Errorf("unique ID -30 gave %+v, %v, want Steve", player, ok)
	}
	e.RemoveByUniqueID(-90)
	if _, ok := e.Get(9); ok {
		t.Error("the zombie is still kept after it was removed")
	}
	e.SetPosition(42, [3]float32{5, 5, 5})
	if _, ok := e.Get(42); ok {
		t.Error("moving an unknown entity made it known")
	}
	e.Forget()
	if len(e.All()) != 0 {
		t.Error("entities are kept after Forget")
	}
}

func TestPastTheLimitNewOnesAreRefused(t *testing.T) {
	e := New()
	for id := uint64(1); id <= MaxEntities+5; id++ {
		e.Add(Entity{RuntimeID: id, UniqueID: int64(id)})
	}
	if n, refused := len(e.All()), e.Refused(); n != MaxEntities || refused != 5 {
		t.Fatalf("kept %d and refused %d, want %d and 5", n, refused, MaxEntities)
	}
	if _, ok := e.Get(1); !ok {
		t.Error("the first entity was dropped: old ones must stay")
	}
	// A known entity can still change when the record is full.
	e.Add(Entity{RuntimeID: 1, UniqueID: 1, Type: "minecraft:pig"})
	if pig, _ := e.Get(1); pig.Type != "minecraft:pig" || e.Refused() != 5 {
		t.Errorf("got %+v and %d refused, want the pig and still 5", pig, e.Refused())
	}
}
