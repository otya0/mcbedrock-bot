// Block names: the server numbers a block by a hash of its name and property
// values (block_hash.go), so the table is made by hashing every state of
// every block in Mojang's list and of the world's custom blocks.
package names

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// blockState is one block state: a name such as "minecraft:wheat" and its
// property values, such as growth = 7, in Mojang's words.
type blockState struct {
	name   string
	states []blockStateValue
}

// blockStateValue is one property of a block state, as text.
type blockStateValue struct {
	property, value string
}

// Block is the name and property values of a block number (the block runtime
// ID in packets): "minecraft:oak_door", {"open_bit": "true", ...}. A number
// not in the table gives "" and no values.
func (n *Names) Block(runtimeID uint32) (name string, states map[string]string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	found, ok := n.blocks[runtimeID]
	if !ok {
		return "", nil
	}
	states = make(map[string]string, len(found.states))
	for _, value := range found.states {
		states[value.property] = value.value
	}
	return found.name, states
}

// SetCustomBlocks adds every state of the world's own blocks (StartGame).
// A custom block's definition lists its properties under "properties", each
// with a "name" and the "enum" of its values (not checked in a real world yet).
func (n *Names) SetCustomBlocks(customBlocks []protocol.BlockEntry) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, custom := range customBlocks {
		properties := map[string]blockProperty{}
		var order []string
		list, _ := custom.Properties["properties"].([]any)
		for _, item := range list {
			definition, _ := item.(map[string]any)
			name, _ := definition["name"].(string)
			values, _ := definition["enum"].([]any)
			if name == "" || len(values) == 0 {
				continue
			}
			property := blockProperty{valueType: blockPropertyString}
			for _, value := range values {
				switch value := value.(type) {
				case byte:
					property.valueType = blockPropertyBool
					property.values = append(property.values, fmt.Sprint(value != 0))
				case int32:
					property.valueType = blockPropertyInt
					property.values = append(property.values, fmt.Sprint(value))
				default:
					property.values = append(property.values, fmt.Sprint(value))
				}
			}
			properties[name] = property
			order = append(order, name)
		}
		n.addBlock(custom.Name, order, properties)
	}
}

// maxStatesOfOneBlock stops a block whose properties would have more
// combinations than any real block (the most in the game is a few thousand).
const maxStatesOfOneBlock = 1 << 16

// addBlock puts every state of one block into the table.
func (n *Names) addBlock(name string, propertyNames []string, properties map[string]blockProperty) {
	values := make([][]string, len(propertyNames))
	combinations := 1
	for i, property := range propertyNames {
		values[i] = properties[property].values
		combinations *= max(len(values[i]), 1)
		if combinations > maxStatesOfOneBlock {
			return
		}
	}
	valueTypeOf := func(property string) blockPropertyValueType {
		return properties[property].valueType
	}
	forEachCombination(values, func(chosen []string) {
		states := make(map[string]string, len(chosen))
		state := blockState{name: name}
		for i, value := range chosen {
			states[propertyNames[i]] = value
			state.states = append(state.states, blockStateValue{propertyNames[i], value})
		}
		n.blocks[blockStateHash(name, states, valueTypeOf)] = state
	})
}

// forEachCombination calls f once for every way of choosing one value from
// each list (the first list changes slowest). With no lists, f is called once.
func forEachCombination(values [][]string, f func(chosen []string)) {
	chosen := make([]string, len(values))
	var walk func(i int)
	walk = func(i int) {
		if i == len(values) {
			f(chosen)
			return
		}
		for _, value := range values[i] {
			chosen[i] = value
			walk(i + 1)
		}
	}
	walk(0)
}
