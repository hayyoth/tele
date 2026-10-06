package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"tele/internal/client"
	"tele/internal/peer"
)

var infoCmd = &cobra.Command{
	Use:   "info <peer>",
	Short: "Show peer information",
	Long: `Display detailed information about a user, chat, or channel.

Peer can be @username, channel#ID, chat#ID, or numeric user ID.`,
	Example: `  tel info @telegram          # Show info about @telegram
  tel info channel#123456    # Show info about channel`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	peerSpec := args[0]

	return client.New(ctx, client.Options{SessionPath: sessionPath}, true, func(ctx context.Context, c *telegram.Client) error {
		api := c.API()
		inputPeer, err := peer.Resolve(ctx, api, peerSpec)
		if err != nil {
			return fmt.Errorf("resolve peer %q: %w", peerSpec, err)
		}

		switch p := inputPeer.(type) {
		case *tg.InputPeerUser:
			users, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUser{
				UserID:     p.UserID,
				AccessHash: p.AccessHash,
			}})
			if err != nil {
				return fmt.Errorf("get user: %w", err)
			}
			if len(users) == 0 {
				return fmt.Errorf("user not found")
			}
			u := users[0].(*tg.User)
			fmt.Fprintf(os.Stdout, "Type: User\n")
			fmt.Fprintf(os.Stdout, "ID: %d\n", u.ID)
			if u.FirstName != "" {
				fmt.Fprintf(os.Stdout, "First Name: %s\n", u.FirstName)
			}
			if u.LastName != "" {
				fmt.Fprintf(os.Stdout, "Last Name: %s\n", u.LastName)
			}
			if u.Username != "" {
				fmt.Fprintf(os.Stdout, "Username: @%s\n", u.Username)
			}
			if u.Phone != "" {
				fmt.Fprintf(os.Stdout, "Phone: %s\n", u.Phone)
			}
			if u.Bot {
				fmt.Fprintf(os.Stdout, "Bot: Yes\n")
			}
			if u.Verified {
				fmt.Fprintf(os.Stdout, "Verified: Yes\n")
			}
			if u.Premium {
				fmt.Fprintf(os.Stdout, "Premium: Yes\n")
			}

		case *tg.InputPeerChat:
			chats, err := api.MessagesGetChats(ctx, []int64{p.ChatID})
			if err != nil {
				return fmt.Errorf("get chat: %w", err)
			}
			chatsSlice := chats.GetChats()
			if len(chatsSlice) == 0 {
				return fmt.Errorf("chat not found")
			}
			chat := chatsSlice[0].(*tg.Chat)
			fmt.Fprintf(os.Stdout, "Type: Chat\n")
			fmt.Fprintf(os.Stdout, "ID: %d\n", chat.ID)
			fmt.Fprintf(os.Stdout, "Title: %s\n", chat.Title)
			fmt.Fprintf(os.Stdout, "Members: %d\n", chat.ParticipantsCount)

		case *tg.InputPeerChannel:
			channels, err := api.ChannelsGetChannels(ctx, []tg.InputChannelClass{&tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			}})
			if err != nil {
				return fmt.Errorf("get channel: %w", err)
			}
			chatsSlice := channels.GetChats()
			if len(chatsSlice) == 0 {
				return fmt.Errorf("channel not found")
			}
			ch := chatsSlice[0].(*tg.Channel)
			fmt.Fprintf(os.Stdout, "Type: Channel\n")
			fmt.Fprintf(os.Stdout, "ID: %d\n", ch.ID)
			fmt.Fprintf(os.Stdout, "Title: %s\n", ch.Title)
			if ch.Username != "" {
				fmt.Fprintf(os.Stdout, "Username: @%s\n", ch.Username)
			}
			if ch.ParticipantsCount > 0 {
				fmt.Fprintf(os.Stdout, "Members: %d\n", ch.ParticipantsCount)
			}
			if ch.Broadcast {
				fmt.Fprintf(os.Stdout, "Broadcast: Yes\n")
			}
			if ch.Megagroup {
				fmt.Fprintf(os.Stdout, "Megagroup: Yes\n")
			}
			if ch.Verified {
				fmt.Fprintf(os.Stdout, "Verified: Yes\n")
			}
		default:
			return fmt.Errorf("unsupported peer type: %T", p)
		}

		return nil
	})
}
