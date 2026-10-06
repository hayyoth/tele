package messages

import (
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// SenderName returns a readable sender name from message and entities.
func SenderName(msg *tg.Message, ent peer.Entities) string {
	if msg.FromID == nil {
		return ""
	}
	switch f := msg.FromID.(type) {
	case *tg.PeerUser:
		if u, ok := ent.User(f.UserID); ok {
			name := strings.TrimSpace(u.FirstName + " " + u.LastName)
			if name == "" {
				name = "(no name)"
			}
			if u.Username != "" {
				return fmt.Sprintf("%s (@%s)", name, u.Username)
			}
			return name
		}
		return fmt.Sprintf("User#%d", f.UserID)
	case *tg.PeerChannel:
		if ch, ok := ent.Channel(f.ChannelID); ok {
			return ch.Title
		}
		return fmt.Sprintf("Channel#%d", f.ChannelID)
	default:
		return "-"
	}
}
