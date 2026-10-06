package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/hayyoth/tele/internal/client"
	"github.com/hayyoth/tele/internal/peer"
)

var forwardCmd = &cobra.Command{
	Use:   "forward <from-peer> <to-peer> <msg-id>",
	Short: "Forward a message",
	Long: `Forward a message from one chat to another.

from-peer: source chat (where the message is)
to-peer: destination chat (where to forward)
msg-id: message ID to forward (use 'tel read --msg-id' to find it)`,
	Example: `  tel forward @user1 @user2 12345     # Forward msg 12345 from @user1 to @user2
  tel forward channel#123 @me 67890    # Forward from channel to saved messages`,
	Args: cobra.ExactArgs(3),
	RunE: runForward,
}

func init() {
	rootCmd.AddCommand(forwardCmd)
}

func runForward(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	fromPeerSpec := args[0]
	toPeerSpec := args[1]
	msgIDStr := args[2]

	var msgID int
	if _, err := fmt.Sscanf(msgIDStr, "%d", &msgID); err != nil {
		return fmt.Errorf("invalid message ID: %q", msgIDStr)
	}

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		fromPeer, err := peer.Resolve(ctx, api, fromPeerSpec)
		if err != nil {
			return fmt.Errorf("resolve from-peer %q: %w", fromPeerSpec, err)
		}
		toPeer, err := peer.Resolve(ctx, api, toPeerSpec)
		if err != nil {
			return fmt.Errorf("resolve to-peer %q: %w", toPeerSpec, err)
		}

		// Find the message to forward
		iter := query.Messages(api).GetHistory(fromPeer).BatchSize(100).Iter()
		var forwardMsg *tg.Message
		for iter.Next(ctx) {
			e := iter.Value()
			if msg, ok := e.Msg.(*tg.Message); ok && msg.ID == msgID {
				forwardMsg = msg
				break
			}
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("search message: %w", err)
		}
		if forwardMsg == nil {
			return fmt.Errorf("message %d not found in %q", msgID, fromPeerSpec)
		}

		// Forward using MessagesForwardMessages API
		_, err = api.MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
			Silent:       false,
			Background:   false,
			WithMyScore: false,
			FromPeer:    fromPeer,
			ID:           []int{forwardMsg.ID},
			ToPeer:       toPeer,
		})
		if err != nil {
			return fmt.Errorf("forward message: %w", err)
		}

		fmt.Fprintln(os.Stdout, "forwarded")
		return nil
	})
}
