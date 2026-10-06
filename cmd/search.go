package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"tele/internal/client"
	"tele/internal/format"
	msghelper "tele/internal/messages"
	"tele/internal/peer"
)

var (
	searchLimit int
	searchFormat string
)

var searchCmd = &cobra.Command{
	Use:   "search <peer> <query>",
	Short: "Search messages",
	Long: `Search for messages containing a specific query in a chat.

Peer can be @username, channel#ID, chat#ID, or numeric user ID.
Query is searched in message text (case-insensitive).`,
	Example: `  tel search @channel "hello"        # Search for "hello" in @channel
  tel search @user "keyword" --limit 20  # Search with limit`,
	Args: cobra.ExactArgs(2),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 20, "Max results to return")
	searchCmd.Flags().StringVar(&searchFormat, "format", "plain", "Output format: plain or tsv")
}

func runSearch(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	peerSpec := args[0]
	queryStr := strings.ToLower(args[1])
	limit := searchLimit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		inputPeer, err := peer.Resolve(ctx, api, peerSpec)
		if err != nil {
			return fmt.Errorf("resolve peer %q: %w", peerSpec, err)
		}

		iter := query.Messages(api).GetHistory(inputPeer).BatchSize(100).Iter()
		count := 0
		found := 0

		for count < limit*10 && iter.Next(ctx) { // Search up to 10x limit messages
			e := iter.Value()
			msg, ok := e.Msg.(*tg.Message)
			if !ok {
				continue
			}
			count++

			text := strings.ToLower(msg.Message)
			if !strings.Contains(text, queryStr) {
				continue
			}

			found++
			if found > limit {
				break
			}

			displayText := format.OneLine(msg.Message)
			if displayText == "" {
				displayText = "(media)"
			}

			senderName := msghelper.SenderName(msg, e.Entities)
			if senderName == "" {
				senderName = "-"
			}

			switch searchFormat {
			case "plain":
				fmt.Fprintf(os.Stdout, "[%d] %s: %s\n", msg.ID, senderName, displayText)
			case "tsv":
				fmt.Fprintf(os.Stdout, "%d\t%s\t%s\n", msg.ID, senderName, displayText)
			default:
				return fmt.Errorf("unknown --format %q (use plain|tsv)", searchFormat)
			}
		}

		if err := iter.Err(); err != nil {
			return err
		}

		if found == 0 {
			fmt.Fprintf(os.Stderr, "No messages found matching %q\n", queryStr)
			return nil
		}

		fmt.Fprintf(os.Stderr, "Found %d message(s)\n", found)
		return nil
	})
}
