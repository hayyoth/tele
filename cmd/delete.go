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
	"tele/internal/client"
	"tele/internal/peer"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <peer> <msg-id>",
	Short: "Delete a message",
	Long: `Delete a message from a chat.

Peer can be @username, channel#ID, chat#ID, or numeric user ID.
msg-id is the message ID to delete (use 'tel read --msg-id' to find it).

Note: You can only delete your own messages.`,
	Example: `  tel delete @user 12345        # Delete message 12345 from @user
  tel delete channel#123 67890  # Delete message in channel`,
	Args: cobra.ExactArgs(2),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	peerSpec := args[0]
	msgIDStr := args[1]

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

		// Verify message exists and is ours
		iter := query.Messages(api).GetHistory(inputPeer).BatchSize(100).Iter()
		var foundMsg *tg.Message
		for iter.Next(ctx) {
			e := iter.Value()
			if msg, ok := e.Msg.(*tg.Message); ok && msg.ID == msgID {
				foundMsg = msg
				break
			}
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("search message: %w", err)
		}
		if foundMsg == nil {
			return fmt.Errorf("message %d not found", msgID)
		}

		// Delete the message
		_, err = api.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
			Revoke: true,
			ID:     []int{msgID},
		})
		if err != nil {
			return fmt.Errorf("delete message: %w", err)
		}

		fmt.Fprintln(os.Stdout, "deleted")
		return nil
	})
}
