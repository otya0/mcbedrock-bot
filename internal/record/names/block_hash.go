package names

import (
	"encoding/binary"
	"hash/fnv"
	"sort"
	"strconv"
)

// How the server numbers a block: the number (the "block runtime ID" in
// the packets, with UseBlockNetworkIDHashes set in StartGame) is the
// FNV-1a 32-bit hash of the block state written as little-endian NBT:
//
//	compound "" { string "name" = the block name,
//	              compound "states" { one tag per property, keys in
//	                                  alphabetical order; a bool is a byte
//	                                  tag, an int an int32 tag, a string a
//	                                  string tag } }
//
// Checked against 22,091 block states of CloudburstMC/Data (2026-10-21):
// all match but minecraft:unknown, which the game numbers 0xFFFFFFFE.
// Checked again in examples/chunks against the numbers the server sent
// (stone, grass_block, wheat growth 7, smoker facing west, ...).

// blockStateHash gives the server's number for a block state. The state
// values are text; valueTypeOf says how each is typed.
func blockStateHash(name string, states map[string]string, valueTypeOf func(property string) blockPropertyValueType) uint32 {
	if name == "minecraft:unknown" {
		return 0xFFFFFFFE
	}
	h := fnv.New32a()
	h.Write(blockStateNBT(name, states, valueTypeOf))
	return h.Sum32()
}

// blockStateNBT writes the block state as the little-endian NBT the hash
// is taken over.
func blockStateNBT(name string, states map[string]string, valueTypeOf func(property string) blockPropertyValueType) []byte {
	const (
		tagEnd      = 0
		tagByte     = 1
		tagInt      = 3
		tagString   = 8
		tagCompound = 10
	)
	var b []byte
	str := func(s string) {
		b = binary.LittleEndian.AppendUint16(b, uint16(len(s)))
		b = append(b, s...)
	}
	b = append(b, tagCompound)
	str("")
	b = append(b, tagString)
	str("name")
	str(name)
	b = append(b, tagCompound)
	str("states")
	keys := make([]string, 0, len(states))
	for k := range states {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := states[k]
		switch valueTypeOf(k) {
		case blockPropertyBool:
			b = append(b, tagByte)
			str(k)
			if v == "true" || v == "1" {
				b = append(b, 1)
			} else {
				b = append(b, 0)
			}
		case blockPropertyInt:
			n, _ := strconv.ParseInt(v, 10, 32)
			b = append(b, tagInt)
			str(k)
			b = binary.LittleEndian.AppendUint32(b, uint32(int32(n)))
		default:
			b = append(b, tagString)
			str(k)
			str(v)
		}
	}
	b = append(b, tagEnd) // end of states
	b = append(b, tagEnd) // end of the root compound
	return b
}
