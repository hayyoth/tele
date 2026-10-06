package peer

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"tele/internal/cache"
)

// Resolve resolves a peer spec (e.g. @username or username) to InputPeerClass.
// Supports:
// - @username / username (resolved via Telegram)
// - cached ids from `convos` output: channel#<id>, chat#<id>, or numeric user id (needs cache for access_hash)
func Resolve(ctx context.Context, api *tg.Client, spec string) (tg.InputPeerClass, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("peer spec is empty")
	}

	// First try cache for non-@ forms.
	isAt := strings.HasPrefix(spec, "@")
	if !isAt {
		cc, err := cache.Load("")
		if err == nil {
			// direct key
			if p, ok, err := cc.Get(spec); err != nil {
				return nil, err
			} else if ok {
				return p, nil
			}
			// numeric key
			if p, ok, err := cc.Get(strings.TrimPrefix(spec, "user#")); err != nil {
				return nil, err
			} else if ok {
				return p, nil
			}
		}
	}

	// Normalize username.
	if strings.HasPrefix(spec, "@") {
		spec = strings.TrimPrefix(spec, "@")
	}

	// If user passed something like channel#123, try resolve by domain without prefix.
	re := regexp.MustCompile(`^(channel|chat|user)#(\d+)$`)
	if m := re.FindStringSubmatch(spec); len(m) == 3 {
		spec = m[2]
	}

	sender := message.NewSender(api)
	req := sender.Resolve(spec)
	return req.AsInputPeer(ctx)
}
