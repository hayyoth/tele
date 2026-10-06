package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"tele/internal/client"
	"tele/internal/config"
)

var (
	loginPhone string
	loginAppID int
	loginAppHash string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to your Telegram account",
	Long: `Authenticate with Telegram and save session.

You'll be prompted for:
  - APP_ID and APP_HASH (if not provided via flags and not in config)
  - Phone number (international format, e.g., +84901234567)
  - Verification code (sent to your Telegram app)
  - 2FA password (if enabled)

APP_ID and APP_HASH are saved to ~/.config/tele/config.json for future use.
Session is saved to ~/.config/tele/session.json for future use.`,
	Example: `  tel login                                    # Interactive login (prompts for APP_ID/HASH)
  tel login --app-id 12345 --app-hash abc123    # Provide credentials
  tel login --phone +84901234567                # Pre-fill phone number`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginPhone, "phone", "p", "", "Phone number (international); if omitted, prompt on stderr")
	loginCmd.Flags().IntVar(&loginAppID, "app-id", 0, "Telegram APP_ID (from https://my.telegram.org/)")
	loginCmd.Flags().StringVar(&loginAppHash, "app-hash", "", "Telegram APP_HASH (from https://my.telegram.org/)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	appID := loginAppID
	appHash := loginAppHash

	// Load existing config to check if credentials are already saved
	cfg, err := config.Load("")
	if err == nil && cfg.IsValid() {
		// Use existing config if flags not provided
		if appID == 0 {
			appID = cfg.AppID
		}
		if appHash == "" {
			appHash = cfg.AppHash
		}
	}

	// Prompt for credentials if not provided and not in config
	if appID == 0 || appHash == "" {
		reader := bufio.NewReader(os.Stdin)
		if appID == 0 {
			fmt.Fprint(os.Stderr, "APP_ID (from https://my.telegram.org/): ")
			idStr, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read APP_ID: %w", err)
			}
			idStr = strings.TrimSpace(idStr)
			appID, err = strconv.Atoi(idStr)
			if err != nil {
				return fmt.Errorf("invalid APP_ID: %w", err)
			}
		}
		if appHash == "" {
			fmt.Fprint(os.Stderr, "APP_HASH (from https://my.telegram.org/): ")
			appHash, err = reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read APP_HASH: %w", err)
			}
			appHash = strings.TrimSpace(appHash)
			if appHash == "" {
				return fmt.Errorf("APP_HASH cannot be empty")
			}
		}

		// Save credentials to config file
		cfg := config.Config{
			AppID:   appID,
			AppHash: appHash,
		}
		if err := config.Save("", cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Credentials saved to config file\n")
		}
	}

	return client.New(ctx, client.Options{
		SessionPath: sessionPath,
		AuthPhone:   loginPhone,
		AppID:       appID,
		AppHash:     appHash,
	}, true, func(ctx context.Context, c *telegram.Client) error {
		self, err := c.Self(ctx)
		if err != nil {
			return err
		}
		// Login success: human message to stderr, username to stdout for scripts
		fmt.Fprintf(os.Stderr, "Logged in: %s %s (@%s)\n", self.FirstName, self.LastName, self.Username)
		fmt.Println(self.Username)
		return nil
	})
}
