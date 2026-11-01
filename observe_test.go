// Checks that the OnTick functions are called in the order they were added,
// with the tick number.
package mcbedrock

import "testing"

func TestOnTickFunctionsInOrder(t *testing.T) {
	b := &Bot{}
	var told []string
	b.OnTick(func(tick uint64) {
		if tick == 3 {
			told = append(told, "first")
		}
	})
	b.OnTick(func(tick uint64) {
		if tick == 3 {
			told = append(told, "second")
		}
	})
	b.tellTick(3)
	if len(told) != 2 || told[0] != "first" || told[1] != "second" {
		t.Fatalf("told %v, want first then second for tick 3", told)
	}
}
