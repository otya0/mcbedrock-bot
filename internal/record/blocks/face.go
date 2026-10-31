// The faces of a block as packets number them, and the point on a face that a
// click names. Also how a position is written in the logs.
package blocks

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// IsFace is true for 0 down, 1 up, 2 north, 3 south, 4 west, 5 east.
func IsFace(face int32) bool {
	return face >= 0 && face <= 5
}

// FaceCentre is the middle of a face as a point within the block, 0 to 1 on
// each side: the up face is {0.5, 1, 0.5}.
func FaceCentre(face int32) [3]float32 {
	centre := [3]float32{0.5, 0.5, 0.5}
	switch face {
	case 0:
		centre[1] = 0
	case 1:
		centre[1] = 1
	case 2:
		centre[2] = 0
	case 3:
		centre[2] = 1
	case 4:
		centre[0] = 0
	case 5:
		centre[0] = 1
	}
	return centre
}

// PositionText writes a position as "x,y,z".
func PositionText(position protocol.BlockPos) string {
	return fmt.Sprintf("%d,%d,%d", position[0], position[1], position[2])
}
