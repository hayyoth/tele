package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/hayyoth/tele/internal/client"
	"github.com/hayyoth/tele/internal/media"
	"github.com/hayyoth/tele/internal/peer"
)

var (
	downloadLimit int
	downloadDir   string
	downloadAll   bool
	downloadMsgID int // download specific message by ID
	downloadIndex  int // download by index (1-based from read output)
)

var downloadCmd = &cobra.Command{
	Use:     "download <peer>",
	Aliases: []string{"dl", "get"},
	Short:   "Download media files",
	Long: `Download media files (photos, videos, documents) from a chat.

By default downloads all media. Use --msg-id or --index to download a specific file.`,
	Example: `  tel download @channel              # Download all media
  tel download @user -d ./media       # Download to ./media
  tel download @user --msg-id 12345   # Download specific message
  tel download @user --index 5        # Download 5th message with media`,
	Args: cobra.ExactArgs(1),
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().IntVarP(&downloadLimit, "limit", "n", 50, "Max messages to scan for media")
	downloadCmd.Flags().StringVarP(&downloadDir, "dir", "d", ".", "Output directory")
	downloadCmd.Flags().BoolVarP(&downloadAll, "all", "a", false, "Download all media (default: only recent)")
	downloadCmd.Flags().IntVar(&downloadMsgID, "msg-id", 0, "Download media from specific message ID")
	downloadCmd.Flags().IntVar(&downloadIndex, "index", 0, "Download media from message at index (1-based, from 'tel read' output)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	peerSpec := args[0]
	limit := downloadLimit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if downloadAll {
		limit = 100 // max per request
	}

	dir := downloadDir
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		inputPeer, err := peer.Resolve(ctx, api, peerSpec)
		if err != nil {
			return fmt.Errorf("resolve peer %q: %w", peerSpec, err)
		}

		dl := downloader.NewDownloader()
		iter := query.Messages(api).GetHistory(inputPeer).BatchSize(limit).Iter()
		count := 0
		downloaded := 0

		// If downloading specific message by ID or index
		if downloadMsgID > 0 || downloadIndex > 0 {
			index := 1
			for iter.Next(ctx) {
				e := iter.Value()
				msg, ok := e.Msg.(*tg.Message)
				if !ok {
					continue
				}
				count++

				// Check if this is the message we want
				matches := false
				if downloadMsgID > 0 && msg.ID == downloadMsgID {
					matches = true
				} else if downloadIndex > 0 && index == downloadIndex {
					matches = true
				}
				index++

				if !matches {
					continue
				}

				if !media.HasFile(msg) {
					return fmt.Errorf("message %d has no media", msg.ID)
				}

			loc, filename, err := media.FileLocation(msg)
			if err != nil {
				return fmt.Errorf("get file location: %w", err)
			}

			filename = filepath.Base(filename)
			if filename == "" || filename == "." {
				filename = fmt.Sprintf("file_%d", msg.ID)
			}
			filename = media.SanitizeFilename(filename)
			outputPath := filepath.Join(dir, filename)

				if _, err := os.Stat(outputPath); err == nil {
					fmt.Fprintf(os.Stderr, "File exists: %s\n", outputPath)
					fmt.Fprintln(os.Stdout, outputPath)
					return nil
				}

				fmt.Fprintf(os.Stderr, "Downloading %s...\n", filename)
				if _, err := dl.Download(api, loc).ToPath(ctx, outputPath); err != nil {
					return fmt.Errorf("download: %w", err)
				}

				fmt.Fprintln(os.Stdout, outputPath)
				return nil
			}
			if err := iter.Err(); err != nil {
				return err
			}
			return fmt.Errorf("message not found")
		}

		// Download all media
		for count < limit && iter.Next(ctx) {
			e := iter.Value()
			msg, ok := e.Msg.(*tg.Message)
			if !ok {
				continue
			}
			count++

			if !media.HasFile(msg) {
				continue
			}

			loc, filename, err := media.FileLocation(msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skip msg %d: %v\n", msg.ID, err)
				continue
			}

			filename = filepath.Base(filename)
			if filename == "" || filename == "." {
				filename = fmt.Sprintf("file_%d", msg.ID)
			}
			filename = media.SanitizeFilename(filename)
			outputPath := filepath.Join(dir, filename)

			if _, err := os.Stat(outputPath); err == nil {
				fmt.Fprintf(os.Stderr, "Skip %s (exists)\n", filename)
				continue
			}

			fmt.Fprintf(os.Stderr, "Downloading %s...\n", filename)
			if _, err := dl.Download(api, loc).ToPath(ctx, outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", filename, err)
				continue
			}

			downloaded++
			fmt.Fprintln(os.Stdout, outputPath)
		}

		if err := iter.Err(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Downloaded %d file(s) to %s\n", downloaded, dir)
		return nil
	})
}
