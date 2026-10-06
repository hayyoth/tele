package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	cliauth "tele/internal/auth"
	"tele/internal/env"
)

// Options for creating the Telegram client.
type Options struct {
	SessionPath string // optional; default from env
	AuthPhone   string // optional; for login, from --phone flag
	AppID       int    // optional; if > 0, use this instead of loading from config/env
	AppHash     string // optional; if not empty, use this instead of loading from config/env
}

// New creates a Telegram client and runs it with flood-wait handling.
// run is called inside client.Run; auth is performed if necessary when needAuth is true.
// run receives the telegram.Client so it can call .API(), .Self(), etc.
func New(ctx context.Context, opt Options, needAuth bool, run func(ctx context.Context, client *telegram.Client) error) error {
	var appID int
	var appHash string
	var err error

	// Use provided AppID/AppHash if available, otherwise load from config/env
	if opt.AppID > 0 && opt.AppHash != "" {
		appID = opt.AppID
		appHash = opt.AppHash
	} else {
		appID, err = env.AppID()
		if err != nil {
			return err
		}
		appHash = env.AppHash()
		if appHash == "" {
			return fmt.Errorf("APP_HASH not set. Run 'tel login --app-id <id> --app-hash <hash>' first or set APP_HASH environment variable")
		}
	}

	sessionPath := opt.SessionPath
	if sessionPath == "" {
		sessionPath = env.SessionPath()
	}
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0700); err != nil {
		return err
	}

	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		fmt.Fprintf(os.Stderr, "FLOOD_WAIT: waiting %v\n", wait.Duration)
	})

	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: sessionPath},
		Middlewares:    []telegram.Middleware{waiter},
	})

	flow := auth.NewFlow(cliauth.Terminal{PhoneNumber: opt.AuthPhone}, auth.SendCodeOptions{})

	return waiter.Run(ctx, func(ctx context.Context) error {
		return client.Run(ctx, func(ctx context.Context) error {
			if needAuth {
				if err := client.Auth().IfNecessary(ctx, flow); err != nil {
					return fmt.Errorf("auth: %w", err)
				}
			}
			return run(ctx, client)
		})
	})
}
