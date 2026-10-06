package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hayyoth/tele/internal/config"
)

// AppID returns APP_ID from config file, then environment variable, then error.
func AppID() (int, error) {
	// Try config file first
	cfg, err := config.Load("")
	if err == nil && cfg.IsValid() {
		return cfg.AppID, nil
	}

	// Fallback to environment variable
	v := os.Getenv("APP_ID")
	if v == "" {
		return 0, fmt.Errorf("APP_ID not set. Run 'tel login --app-id <id> --app-hash <hash>' first or set APP_ID environment variable")
	}
	id, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("APP_ID must be a number: %w", err)
	}
	return id, nil
}

// AppHash returns APP_HASH from config file, then environment variable, then empty string.
func AppHash() string {
	// Try config file first
	cfg, err := config.Load("")
	if err == nil && cfg.IsValid() {
		return cfg.AppHash
	}

	// Fallback to environment variable
	return os.Getenv("APP_HASH")
}

// SessionPath returns SESSION_PATH from environment, or default.
func SessionPath() string {
	if p := os.Getenv("SESSION_PATH"); p != "" {
		return p
	}
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "tele", "session.json")
}

// DialogsPageSize returns DIALOGS_PAGE_SIZE from environment, or default 10.
func DialogsPageSize() int {
	v := os.Getenv("DIALOGS_PAGE_SIZE")
	if v == "" {
		return 10
	}
	size, err := strconv.Atoi(v)
	if err != nil || size <= 0 {
		return 10
	}
	if size > 100 {
		return 100
	}
	return size
}
