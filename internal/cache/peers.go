package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/td/tg"
)

// PeerCache stores a minimal mapping from a human id (like channel#123) to InputPeer.
type PeerCache struct {
	Peers map[string]PeerEntry `json:"peers"`
}

type PeerEntry struct {
	Type       string `json:"type"` // user|chat|channel
	ID         int64  `json:"id"`
	AccessHash int64  `json:"access_hash,omitempty"`
}

func DefaultPath() string {
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "tele", "peers.json")
}

func Load(path string) (PeerCache, error) {
	if path == "" {
		path = DefaultPath()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PeerCache{Peers: map[string]PeerEntry{}}, nil
		}
		return PeerCache{}, err
	}
	var c PeerCache
	if err := json.Unmarshal(b, &c); err != nil {
		return PeerCache{}, err
	}
	if c.Peers == nil {
		c.Peers = map[string]PeerEntry{}
	}
	return c, nil
}

func Save(path string, c PeerCache) error {
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func NormalizeKey(k string) string {
	k = strings.TrimSpace(k)
	k = strings.TrimPrefix(k, "@")
	return k
}

func (c *PeerCache) Put(key string, p tg.InputPeerClass) {
	if c.Peers == nil {
		c.Peers = map[string]PeerEntry{}
	}
	key = NormalizeKey(key)
	switch v := p.(type) {
	case *tg.InputPeerUser:
		c.Peers[key] = PeerEntry{Type: "user", ID: v.UserID, AccessHash: v.AccessHash}
	case *tg.InputPeerChat:
		c.Peers[key] = PeerEntry{Type: "chat", ID: v.ChatID}
	case *tg.InputPeerChannel:
		c.Peers[key] = PeerEntry{Type: "channel", ID: v.ChannelID, AccessHash: v.AccessHash}
	}
}

func (c PeerCache) Get(key string) (tg.InputPeerClass, bool, error) {
	key = NormalizeKey(key)
	e, ok := c.Peers[key]
	if !ok {
		return nil, false, nil
	}
	switch e.Type {
	case "user":
		return &tg.InputPeerUser{UserID: e.ID, AccessHash: e.AccessHash}, true, nil
	case "chat":
		return &tg.InputPeerChat{ChatID: e.ID}, true, nil
	case "channel":
		return &tg.InputPeerChannel{ChannelID: e.ID, AccessHash: e.AccessHash}, true, nil
	default:
		return nil, false, fmt.Errorf("unknown cached peer type %q", e.Type)
	}
}

