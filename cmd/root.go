package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	sessionPath string
)

var rootCmd = &cobra.Command{
	Use:   "tel",
	Short: "Telegram CLI client",
	Long: `Telegram CLI - Command-line interface for Telegram.

Quick start:
  tel login              # Login to your account
  tel chats             # List conversations
  tel read @username    # Read messages from a chat
  tel send @user "hi"   # Send a message

For more information, use 'tel <command> --help'`,
	Example: `  tel login
  tel chats --preview
  tel read @username --limit 20
  tel send @user "Hello!"
  tel send @user --file photo.jpg "Check this out"
  tel download @channel`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&sessionPath, "session", "", "Session file path (default: ~/.config/tele/session.json)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
