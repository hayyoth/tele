package cmd

import (
	"context"
	"fmt"
	"mime"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/hayyoth/tele/internal/client"
	"github.com/hayyoth/tele/internal/peer"
)

var (
	sendFile string
	sendAs   string // "auto", "document", "photo", "video"
)

var sendCmd = &cobra.Command{
	Use:     "send <peer> [text]",
	Aliases: []string{"msg", "text"},
	Short:   "Send a message",
	Long: `Send a text message or file to a chat.

Peer can be @username, channel#ID, or numeric ID.
If text is omitted, only file will be sent (if --file is provided).

Use --as to specify how to send the file:
  - auto (default): Auto-detect from file type
  - document: Send as document
  - photo: Send as photo
  - video: Send as video`,
	Example: `  tel send @user "Hello!"                    # Send text
  tel send @user --file photo.jpg              # Send file (auto-detect)
  tel send @user --file pic.jpg "Check"        # Send file with caption
  tel send @user --file image.jpg --as photo    # Force send as photo
  tel send @user --file video.mp4 --as video    # Force send as video
  tel send channel#123 "Announcement"          # Send to channel`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().StringVarP(&sendFile, "file", "f", "", "Path to file to send (optional)")
	sendCmd.Flags().StringVar(&sendAs, "as", "auto", "How to send file: auto, document, photo, or video")
}

func runSend(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	peerSpec := args[0]
	text := ""
	if len(args) > 1 {
		text = args[1]
	}

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		inputPeer, err := peer.Resolve(ctx, api, peerSpec)
		if err != nil {
			return fmt.Errorf("resolve peer %q: %w", peerSpec, err)
		}
		up := uploader.NewUploader(api)
		sender := message.NewSender(api).WithUploader(up).To(inputPeer)

		if sendFile != "" {
			// Get filename and extension
			filename := filepath.Base(sendFile)
			ext := strings.ToLower(filepath.Ext(filename))
			
			// Detect mime type from extension
			mimeType := mime.TypeByExtension(ext)
			if mimeType == "" {
				// Fallback to application/octet-stream if unknown
				mimeType = "application/octet-stream"
			}
			
			// Upload file
			file, err := up.FromPath(ctx, sendFile)
			if err != nil {
				return fmt.Errorf("upload file: %w", err)
			}
			
			// Determine how to send the file
			asType := strings.ToLower(sendAs)
			if asType == "auto" {
				// Auto-detect from mime type
				if strings.HasPrefix(mimeType, "image/") {
					asType = "photo"
				} else if strings.HasPrefix(mimeType, "video/") {
					asType = "video"
				} else {
					asType = "document"
				}
			}
			
			// Send based on type
			switch asType {
			case "photo":
				photo := message.UploadedPhoto(file)
				_, err = sender.Media(ctx, photo)
				if err == nil && text != "" {
					// Send caption as separate message
					_, _ = sender.Text(ctx, text)
				}
			case "video":
				video := message.UploadedDocument(file).
					Filename(filename).
					MIME(mimeType).
					Video()
				_, err = sender.Media(ctx, video)
				if err == nil && text != "" {
					// Send caption as separate message
					_, _ = sender.Text(ctx, text)
				}
			default: // "document"
				doc := message.UploadedDocument(file).
					Filename(filename).
					MIME(mimeType)
				_, err = sender.Media(ctx, doc)
				if err == nil && text != "" {
					// Send caption as separate message
					_, _ = sender.Text(ctx, text)
				}
			}
			
			if err != nil {
				return fmt.Errorf("send file: %w", err)
			}
			// Success: id or ok to stdout for scripts
			fmt.Fprintln(os.Stdout, "sent")
			return nil
		}
		if text == "" {
			return fmt.Errorf("provide message text or --file")
		}
		_, err = sender.Text(ctx, text)
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}
		fmt.Fprintln(os.Stdout, "sent")
		return nil
	})
}
