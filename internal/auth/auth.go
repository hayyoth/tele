package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// Terminal implements auth.UserAuthenticator using stdin (no phone in env).
type Terminal struct {
	PhoneNumber string // optional, from --phone flag; if empty, prompt
}

func (t Terminal) Phone(ctx context.Context) (string, error) {
	if t.PhoneNumber != "" {
		return t.PhoneNumber, nil
	}
	fmt.Fprint(os.Stderr, "Phone (international, e.g. +84...): ")
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (t Terminal) Password(ctx context.Context) (string, error) {
	fmt.Fprint(os.Stderr, "2FA password (or Enter to skip): ")
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (t Terminal) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (t Terminal) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("sign-up not supported")
}

func (t Terminal) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Fprint(os.Stderr, "Code from Telegram: ")
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
