package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/hayyoth/tele/internal/client"
	"github.com/hayyoth/tele/internal/cache"
	"github.com/hayyoth/tele/internal/dialogs"
	"github.com/hayyoth/tele/internal/env"
	"github.com/hayyoth/tele/internal/format"
)

var (
	convosLimit   int
	convosPage    int
	convosHideID  bool
	convosPreview bool
	convosFormat  string
	convosMore  bool
	convosMeta    bool
)

var convosCmd = &cobra.Command{
	Use:     "chats",
	Aliases: []string{"convos", "ls", "list"},
	Short:   "List conversations",
	Long: `List all your Telegram conversations (chats, channels, groups).

By default, shows conversation titles with peer IDs (10 per page).
Press Enter to load more conversations (use --more to pagination).`,
	Example: `  tel chats                    # List with IDs (default, 10 per page)
  tel chats --page 2           # Jump to page 2
  tel chats --limit 20         # 20 per page
  tel chats --meta             # Show pagination info
  tel chats --preview          # Show last message preview
  tel chats --more             # Show pagination`,
	RunE: runConvos,
}

func init() {
	rootCmd.AddCommand(convosCmd)
	convosCmd.Flags().IntVarP(&convosLimit, "limit", "n", 0, "Conversations per page (default: 10)")
	convosCmd.Flags().IntVarP(&convosPage, "page", "p", 1, "Jump to specific page (1-based)")
	convosCmd.Flags().BoolVar(&convosHideID, "no-id", false, "Hide peer IDs (IDs are shown by default)")
	convosCmd.Flags().BoolVar(&convosPreview, "preview", false, "Show last message preview")
	convosCmd.Flags().BoolVar(&convosMeta, "meta", false, "Show pagination info (page, total)")
	convosCmd.Flags().BoolVar(&convosMore, "more", false, "Prompt for 'load more'")
	convosCmd.Flags().StringVar(&convosFormat, "format", "table", "Output format: table (human) or tsv (script-friendly)")
}

func runConvos(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	limit := convosLimit
	if limit <= 0 {
		limit = env.DialogsPageSize() // default 10
	}
	if limit > 100 {
		limit = 100
	}
	pageSize := limit
	targetPage := convosPage
	if targetPage < 1 {
		targetPage = 1
	}

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		iter := query.GetDialogs(api).BatchSize(pageSize).Iter()
		peerCache, _ := cache.Load("")
		showID := !convosHideID
		currentPage := 1
		totalShown := 0

		// Skip to target page
		if targetPage > 1 {
			skipCount := (targetPage - 1) * pageSize
			for skipCount > 0 && iter.Next(ctx) {
				e := iter.Value()
				if !e.Deleted() {
					skipCount--
				}
			}
			currentPage = targetPage
		}

		for {
			itemsInPage := 0

			// Header only on first page
			if convosFormat == "table" && currentPage == 1 {
				if showID && convosPreview {
					_, _ = fmt.Fprintln(os.Stdout, "#  ID\tTITLE\tLAST")
				} else if showID {
					_, _ = fmt.Fprintln(os.Stdout, "#  ID\tTITLE")
				} else if convosPreview {
					_, _ = fmt.Fprintln(os.Stdout, "#  TITLE\tLAST")
				} else {
					_, _ = fmt.Fprintln(os.Stdout, "#  TITLE")
				}
			}

			// Show page metadata
			if convosMeta {
				fmt.Fprintf(os.Stderr, "Page %d", currentPage)
			}

			// Fetch items for current page
			for itemsInPage < pageSize && iter.Next(ctx) {
				e := iter.Value()
				if e.Deleted() {
					continue
				}
				// Cache peer
				if e.Peer != nil {
					key := dialogs.PeerID(e)
					if key != "" {
						peerCache.Put(key, e.Peer)
					}
					switch p := e.Peer.(type) {
					case *tg.InputPeerUser:
						peerCache.Put(fmt.Sprintf("%d", p.UserID), e.Peer)
					}
				}
				itemsInPage++
				totalShown++
				id := dialogs.PeerID(e)
				title := format.Trunc(format.OneLine(dialogs.Title(e)), 48)
				preview := format.Trunc(format.OneLine(dialogs.LastPreview(e)), 60)

				switch convosFormat {
				case "tsv":
					if showID && convosPreview {
						_, _ = fmt.Fprintf(os.Stdout, "%s\t%s\t%s\n", id, title, preview)
					} else if showID {
						_, _ = fmt.Fprintf(os.Stdout, "%s\t%s\n", id, title)
					} else if convosPreview {
						_, _ = fmt.Fprintf(os.Stdout, "%s\t%s\n", title, preview)
					} else {
						_, _ = fmt.Fprintln(os.Stdout, title)
					}
				case "table":
					if showID && convosPreview {
						_, _ = fmt.Fprintf(os.Stdout, "%-3d\t%s\t%s\t%s\n", totalShown, id, title, preview)
					} else if showID {
						_, _ = fmt.Fprintf(os.Stdout, "%-3d\t%s\t%s\n", totalShown, id, title)
					} else if convosPreview {
						_, _ = fmt.Fprintf(os.Stdout, "%-3d\t%s\t%s\n", totalShown, title, preview)
					} else {
						_, _ = fmt.Fprintf(os.Stdout, "%-3d\t%s\n", totalShown, title)
					}
				}
			}

			if err := iter.Err(); err != nil {
				return err
			}

			// Check if there's more (full page = likely more)
			hasMore := itemsInPage == pageSize

			// Show metadata
			if convosMeta {
				if itemsInPage == 0 {
					fmt.Fprintf(os.Stderr, " (no items)\n")
				} else if hasMore {
					fmt.Fprintf(os.Stderr, " (showing %d, more available)\n", itemsInPage)
				} else {
					fmt.Fprintf(os.Stderr, " (showing %d, end)\n", itemsInPage)
				}
			}

			if itemsInPage == 0 {
				break
			}

			if !convosMore {
				break
			}

			// Prompt for next page
			if hasMore {
				fmt.Fprintf(os.Stderr, "\nPage %d (%d items). Press Enter for next page (Ctrl+C to exit)...", currentPage, itemsInPage)
				reader := bufio.NewReader(os.Stdin)
				if _, err := reader.ReadString('\n'); err != nil {
					break
				}
				currentPage++
			} else {
				break
			}
		}
		_ = cache.Save("", peerCache)
		return nil
	})
}
