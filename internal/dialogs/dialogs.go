package dialogs

import (
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

// Elem is alias for query/dialogs.Elem.
type Elem = dialogs.Elem

// Title returns a display title for a dialog element.
func Title(e dialogs.Elem) string {
	ent := e.Entities
	pclass := e.Dialog.GetPeer()
	if pclass == nil {
		return "(unknown)"
	}
	switch p := pclass.(type) {
	case *tg.PeerUser:
		if u, ok := ent.User(p.UserID); ok {
			name := strings.TrimSpace(u.FirstName + " " + u.LastName)
			if name == "" {
				name = "(no name)"
			}
			if u.Username != "" {
				return name + " @" + u.Username
			}
			return name
		}
		return fmt.Sprintf("User#%d", p.UserID)
	case *tg.PeerChat:
		if c, ok := ent.Chat(p.ChatID); ok {
			return c.Title
		}
		return fmt.Sprintf("Chat#%d", p.ChatID)
	case *tg.PeerChannel:
		if ch, ok := ent.Channel(p.ChannelID); ok {
			return ch.Title
		}
		return fmt.Sprintf("Channel#%d", p.ChannelID)
	default:
		return "(unknown)"
	}
}

// LastPreview returns a short preview of the last message.
func LastPreview(e dialogs.Elem) string {
	last := e.Last
	if msg, ok := last.(*tg.Message); ok && msg.Message != "" {
		s := msg.Message
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	}
	return "(media)"
}

// PeerID returns a stable string id for the peer (for use in messages command).
func PeerID(e dialogs.Elem) string {
	ent := e.Entities
	pclass := e.Dialog.GetPeer()
	if pclass == nil {
		return ""
	}
	switch p := pclass.(type) {
	case *tg.PeerUser:
		if u, ok := ent.User(p.UserID); ok && u.Username != "" {
			return "@" + u.Username
		}
		return fmt.Sprintf("%d", p.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("chat#%d", p.ChatID)
	case *tg.PeerChannel:
		if ch, ok := ent.Channel(p.ChannelID); ok && ch.Username != "" {
			return "@" + ch.Username
		}
		return fmt.Sprintf("channel#%d", p.ChannelID)
	default:
		return ""
	}
}
