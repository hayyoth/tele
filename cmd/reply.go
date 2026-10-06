package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/hayyoth/tele/internal/client"
	"github.com/hayyoth/tele/internal/peer"
)

var replyCmd = &cobra.Command{
	Use:   "reply <peer> <msg-id> <text>",
	Short: "Reply to a message",
	Long: `Reply to a specific message in a chat.

Peer can be @username, channel#ID, chat#ID, or numeric user ID.
msg-id is the message ID to reply to (use 'tel read --msg-id' to find it).`,
	Example: `  tel reply @user 12345 "Thanks!"     # Reply to message 12345
  tel reply channel#123 67890 "Got it"  # Reply in channel`,
	Args: cobra.ExactArgs(3),
	RunE: runReply,
}

func init() {
	rootCmd.AddCommand(replyCmd)
}

func runReply(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	peerSpec := args[0]
	msgIDStr := args[1]
	text := args[2]

	var msgID int
	if _, err := fmt.Sscanf(msgIDStr, "%d", &msgID); err != nil {
		return fmt.Errorf("invalid message ID: %q", msgIDStr)
	}

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		inputPeer, err := peer.Resolve(ctx, api, peerSpec)
		if err != nil {
			return fmt.Errorf("resolve peer %q: %w", peerSpec, err)
		}

		// Find the message to reply to
		iter := query.Messages(api).GetHistory(inputPeer).BatchSize(100).Iter()
		var replyToMsg *tg.Message
		for iter.Next(ctx) {
			e := iter.Value()
			if msg, ok := e.Msg.(*tg.Message); ok && msg.ID == msgID {
				replyToMsg = msg
				break
			}
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("search message: %w", err)
		}
		if replyToMsg == nil {
			return fmt.Errorf("message %d not found", msgID)
		}

		sender := message.NewSender(api).To(inputPeer)
		_, err = sender.Reply(replyToMsg.ID).Text(ctx, text)
		if err != nil {
			return fmt.Errorf("send reply: %w", err)
		}

		fmt.Fprintln(os.Stdout, "sent")
		return nil
	})
}
