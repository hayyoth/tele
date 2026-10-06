package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/hayyoth/tele/internal/client"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user information",
	Long:  "Display information about the currently logged-in Telegram account.",
	Example: `  tel whoami              # Show current user info`,
	RunE: runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	return client.New(ctx, client.Options{SessionPath: sessionPath}, false, func(ctx context.Context, c *telegram.Client) error {
		self, err := c.Self(ctx)
		if err != nil {
			return fmt.Errorf("get self: %w", err)
		}

		fmt.Fprintf(os.Stdout, "ID: %d\n", self.ID)
		fmt.Fprintf(os.Stdout, "First Name: %s\n", self.FirstName)
		if self.LastName != "" {
			fmt.Fprintf(os.Stdout, "Last Name: %s\n", self.LastName)
		}
		if self.Username != "" {
			fmt.Fprintf(os.Stdout, "Username: @%s\n", self.Username)
		}
		if self.Phone != "" {
			fmt.Fprintf(os.Stdout, "Phone: %s\n", self.Phone)
		}
		return nil
	})
}
