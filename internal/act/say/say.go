// Package say sends a chat message (T). A text that starts with "/" goes as
// chat too: the game's own server reads it as chat, some other servers as a
// command. Whether to send one is the caller's choice.
package say

import (
	"errors"
	"fmt"
	"log/slog"
	"unicode/utf8"

	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// maxLength is the most characters the game's chat box takes.
const maxLength = 512

type Say struct {
	link   *serverlink.Link
	logger *slog.Logger
}

func New(link *serverlink.Link, logger *slog.Logger) *Say {
	return &Say{link: link, logger: logger}
}

func (s *Say) Message(text string) error {
	if text == "" {
		return errors.New("say message: the text is empty")
	}
	if length := utf8.RuneCountInString(text); length > maxLength {
		return fmt.Errorf("say message: %d characters is more than %d", length, maxLength)
	}
	identity := s.link.IdentityData()
	s.logger.Info("act", "func", "Say.Message", "text", text, "packet", "text/Chat")
	return s.link.SendText(&packet.Text{TextType: packet.TextTypeChat, SourceName: identity.DisplayName, Message: text, XUID: identity.XUID})
}
