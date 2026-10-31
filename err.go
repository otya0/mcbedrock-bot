// The errors a user can check with errors.Is.
package mcbedrock

import (
	"github.com/otya0/mcbedrock-bot/internal/join"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
)

// ErrWorldNotFound: Join found no world on the LAN. ErrClosed: the reason
// in Err when Close ended the connection.
var (
	ErrWorldNotFound = join.ErrWorldNotFound
	ErrClosed        = serverlink.ErrClosed
)
