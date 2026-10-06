package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/hayyoth/tele/internal/client"
	"github.com/hayyoth/tele/internal/format"
	"github.com/hayyoth/tele/internal/media"
	msghelper "github.com/hayyoth/tele/internal/messages"
	"github.com/hayyoth/tele/internal/peer"
)

var (
	messagesLimit      int
	messagesPage       int
	messagesHideSender bool
	messagesShowID     bool
	messagesFormat     string
	messagesMediaInfo  bool
	messagesMore     bool
	messagesMeta       bool
	messagesSince      string // date filter: YYYY-MM-DD or relative like "7d", "1w", "1m"
	messagesUntil      string // date filter: YYYY-MM-DD
)

var messagesCmd = &cobra.Command{
	Use:     "read <peer>",
	Aliases: []string{"messages", "show", "view"},
	Short:   "Read messages from a chat",
	Long: `Read messages from a conversation.

Peer can be:
  - @username (e.g., @telegram)
  - channel#ID or chat#ID (from 'tel chats')
  - Numeric user ID (if cached)

By default, shows sender name and message text (10 per page).
Press Enter to load more messages (use --more to pagination).`,
	Example: `  tel read @telegram              # Read with sender names (default, 10 per page)
  tel read @telegram --page 2      # Jump to page 2
  tel read @telegram --limit 20    # 20 messages per page
  tel read @telegram --meta        # Show pagination info
  tel read @telegram --msg-id      # Show message IDs
  tel read @telegram --more        # Show pagination`,
	Args: cobra.ExactArgs(1),
	RunE: runMessages,
}

func init() {
	rootCmd.AddCommand(messagesCmd)
	messagesCmd.Flags().IntVarP(&messagesLimit, "limit", "n", 0, "Messages per page (default: 10)")
	messagesCmd.Flags().IntVarP(&messagesPage, "page", "p", 1, "Jump to specific page (1-based)")
	messagesCmd.Flags().BoolVar(&messagesHideSender, "no-sender", false, "Hide sender info (sender is shown by default)")
	messagesCmd.Flags().BoolVar(&messagesShowID, "msg-id", true, "Show message ID (useful for 'tel download --msg-id')")
	messagesCmd.Flags().BoolVar(&messagesMediaInfo, "media-info", false, "Show media details instead of '(media)'")
	messagesCmd.Flags().BoolVar(&messagesMeta, "meta", false, "Show pagination info (page, total)")
	messagesCmd.Flags().BoolVar(&messagesMore, "more", false, "Prompt for 'load more'")
	messagesCmd.Flags().StringVar(&messagesFormat, "format", "plain", "Output format: plain (text only) or tsv (with metadata)")
	messagesCmd.Flags().StringVar(&messagesSince, "since", "", "Show messages since date (YYYY-MM-DD or relative: 7d, 1w, 1m)")
	messagesCmd.Flags().StringVar(&messagesUntil, "until", "", "Show messages until date (YYYY-MM-DD)")
}

func runMessages(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	peerSpec := args[0]
	limit := messagesLimit
	if limit <= 0 {
		limit = 10 // default 10
	}
	if limit > 100 {
		limit = 100
	}
	pageSize := limit
	targetPage := messagesPage
	if targetPage < 1 {
		targetPage = 1
	}

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		inputPeer, err := peer.Resolve(ctx, api, peerSpec)
		if err != nil {
			return fmt.Errorf("resolve peer %q: %w", peerSpec, err)
		}

		// Parse date filters
		var sinceTime, untilTime time.Time
		if messagesSince != "" {
			sinceTime, err = parseDateFilter(messagesSince)
			if err != nil {
				return fmt.Errorf("invalid --since date: %w", err)
			}
		}
		if messagesUntil != "" {
			untilTime, err = time.Parse("2006-01-02", messagesUntil)
			if err != nil {
				return fmt.Errorf("invalid --until date (use YYYY-MM-DD): %w", err)
			}
			// Set to end of day
			untilTime = untilTime.Add(24*time.Hour - time.Second)
		}

		iter := query.Messages(api).GetHistory(inputPeer).BatchSize(pageSize).Iter()
		showSender := !messagesHideSender
		currentPage := 1
		totalShown := 0

		// Skip to target page
		if targetPage > 1 {
			skipCount := (targetPage - 1) * pageSize
			for skipCount > 0 && iter.Next(ctx) {
				e := iter.Value()
				_, ok := e.Msg.(*tg.Message)
				if ok {
					skipCount--
				}
			}
			currentPage = targetPage
		}

		for {
			itemsInPage := 0

			// Show page metadata
			if messagesMeta {
				fmt.Fprintf(os.Stderr, "Page %d", currentPage)
			}

			// Fetch items for current page
			for itemsInPage < pageSize && iter.Next(ctx) {
				e := iter.Value()
				msg, ok := e.Msg.(*tg.Message)
				if !ok {
					continue
				}

				// Apply date filters
				msgTime := time.Unix(int64(msg.Date), 0)
				if !sinceTime.IsZero() && msgTime.Before(sinceTime) {
					continue
				}
				if !untilTime.IsZero() && msgTime.After(untilTime) {
					continue
				}

				itemsInPage++
				totalShown++
				text := msg.Message
				mediaInfo := ""
				
				// Always show media info if message has media
				if msg.Media != nil {
					mediaInfo = media.Info(msg)
					if mediaInfo == "" {
						mediaInfo = "[media]"
					}
				}
				
				// Combine text and media info
				if text != "" && mediaInfo != "" {
					text = text + " " + mediaInfo
				} else if text == "" && mediaInfo != "" {
					text = mediaInfo
				} else if text == "" {
					text = "(no content)"
				}
				
				text = format.OneLine(text)

				// Get sender name
				senderName := ""
				if showSender {
					senderName = msghelper.SenderName(msg, e.Entities)
					if senderName == "" {
						senderName = "-"
					}
				}

				switch messagesFormat {
				case "plain":
					if messagesShowID && showSender {
						_, err := fmt.Fprintf(os.Stdout, "[%d] %s: %s\n", msg.ID, senderName, text)
						if err != nil {
							return err
						}
					} else if messagesShowID {
						_, err := fmt.Fprintf(os.Stdout, "[%d] %s\n", msg.ID, text)
						if err != nil {
							return err
						}
					} else if showSender {
						_, err := fmt.Fprintf(os.Stdout, "%s: %s\n", senderName, text)
						if err != nil {
							return err
						}
					} else {
						_, err := fmt.Fprintln(os.Stdout, text)
						if err != nil {
							return err
						}
					}
				case "tsv":
					date := time.Unix(int64(msg.Date), 0).Format(time.RFC3339)
					if messagesShowID && showSender {
						_, err := fmt.Fprintf(os.Stdout, "%d\t%s\t%s\t%s\n", msg.ID, date, senderName, text)
						if err != nil {
							return err
						}
					} else if messagesShowID {
						_, err := fmt.Fprintf(os.Stdout, "%d\t%s\t%s\n", msg.ID, date, text)
						if err != nil {
							return err
						}
					} else if showSender {
						_, err := fmt.Fprintf(os.Stdout, "%s\t%s\t%s\n", date, senderName, text)
						if err != nil {
							return err
						}
					} else {
						_, err := fmt.Fprintf(os.Stdout, "%s\t%s\n", date, text)
						if err != nil {
							return err
						}
					}
				default:
					return fmt.Errorf("unknown --format %q (use plain|tsv)", messagesFormat)
				}
			}

			if err := iter.Err(); err != nil {
				return err
			}

			// Check if there's more
			hasMore := itemsInPage == pageSize

			// Show metadata
			if messagesMeta {
				if itemsInPage == 0 {
					fmt.Fprintf(os.Stderr, " (no messages)\n")
				} else if hasMore {
					fmt.Fprintf(os.Stderr, " (showing %d, more available)\n", itemsInPage)
				} else {
					fmt.Fprintf(os.Stderr, " (showing %d, end)\n", itemsInPage)
				}
			}

			if itemsInPage == 0 {
				break
			}

			if !messagesMore {
				break
			}

			// Prompt for next page
			if hasMore {
				fmt.Fprintf(os.Stderr, "\nPage %d (%d messages). Press Enter for next page (Ctrl+C to exit)...", currentPage, itemsInPage)
				reader := bufio.NewReader(os.Stdin)
				if _, err := reader.ReadString('\n'); err != nil {
					break
				}
				currentPage++
			} else {
				break
			}
		}
		return nil
	})
}

// parseDateFilter parses relative dates like "7d", "1w", "1m" or absolute dates "YYYY-MM-DD"
func parseDateFilter(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	// Try relative date first
	if len(s) > 1 && (s[len(s)-1] == 'd' || s[len(s)-1] == 'w' || s[len(s)-1] == 'm') {
		numStr := s[:len(s)-1]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid number: %q", numStr)
		}
		now := time.Now()
		switch s[len(s)-1] {
		case 'd':
			return now.AddDate(0, 0, -num), nil
		case 'w':
			return now.AddDate(0, 0, -num*7), nil
		case 'm':
			return now.AddDate(0, -num, 0), nil
		}
	}

	// Try absolute date
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format (use YYYY-MM-DD or relative like 7d, 1w, 1m): %w", err)
	}
	return t, nil
}
