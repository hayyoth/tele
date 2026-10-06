# tele

A powerful command-line interface for Telegram built with Go. Interact with Telegram from your terminal with support for messages, media, and more.

## Features

- 📥 **Media Download** — Download photos, videos, documents, and other media directly from Telegram.
  - 🔓 **Works even when media downloads are restricted by the channel** — retrieve and save media that is otherwise blocked from downloading through Telegram's client.
- 🔐 **Secure Authentication** - Login once, use forever
- 💬 **Send & Receive Messages** - Text messages with full support
- 📁 **File Management** - Send files as documents, photos, or videos
- 🔍 **Search Messages** - Find messages by keyword
- 📋 **List Conversations** - Browse chats, channels, and groups
- 🔄 **Message Operations** - Reply, forward, delete messages
- 📊 **Pagination** - Efficient pagination with interactive "load more"
- 🎯 **Peer Resolution** - Support for @username, channel#ID, chat#ID, and numeric IDs
- 📅 **Date Filtering** - Filter messages by date range
- 🎨 **Multiple Output Formats** - Human-readable tables or machine-friendly TSV

## Installation

### Prerequisites

- Go 1.25.5 or later
- Telegram API credentials (APP_ID and APP_HASH from [my.telegram.org](https://my.telegram.org/))

### Install via Go

```bash
go install github.com/hayyoth/tele@latest
```

Make sure `$GOPATH/bin` or `$GOBIN` is in your `PATH`.

### Build from Source

```bash
git clone <repository-url>
cd tele
go build .
```

The binary `tele` will be created in the current directory.

### Install Globally

```bash
go install .
```

## Quick Start

1. **Get Telegram API credentials:**
   - Go to [https://my.telegram.org/](https://my.telegram.org/)
   - Log in with your phone number
   - Go to "API development tools"
   - Create an application to get `APP_ID` and `APP_HASH`

2. **Login:**
   ```bash
   tele login
   ```
   You'll be prompted for:
   - APP_ID and APP_HASH (first time only, saved automatically)
   - Phone number (international format, e.g., +84901234567)
   - Verification code (sent to your Telegram app)
   - 2FA password (if enabled)

3. **Start using:**
   ```bash
   tele chats              # List your conversations
   tele read @telegram     # Read messages from a chat
   tele send @user "Hi!"   # Send a message
   ```

## Configuration

Credentials are automatically saved to `~/.config/tele/config.json` after first login. You can also provide them via command-line flags:

```bash
tele login --app-id 12345 --app-hash abc123def456
```

### Optional: Environment Variables

You can also set credentials via environment variables (for backward compatibility):

```bash
export APP_ID=12345
export APP_HASH=your_api_hash_here
```

## Commands

### Authentication

#### `tele login`
Log in to your Telegram account.

```bash
tele login                                    # Interactive login
tele login --app-id 12345 --app-hash abc123  # With credentials
tele login --phone +84901234567              # Pre-fill phone
```

**First-time setup:** You'll be prompted for APP_ID and APP_HASH, which are saved automatically.

### Account Information

#### `tele whoami`
Show current user information.

```bash
tele whoami
```

Output:
```
ID: 123456789
First Name: John
Last Name: Doe
Username: @johndoe
Phone: +84901234567
```

### Conversations

#### `tele chats` (aliases: `convos`, `ls`, `list`)
List all your Telegram conversations (chats, channels, groups).

```bash
tele chats                    # List with IDs (default, 10 per page)
tele chats --page 2           # Jump to page 2
tele chats --limit 20         # 20 per page
tele chats --meta             # Show pagination info
tele chats --preview          # Show last message preview
tele chats --no-more          # Disable interactive pagination
tele chats --format tsv       # Machine-readable format
tele chats --no-id            # Hide peer IDs
```

**Features:**
- Pagination with "Press Enter for next page"
- Shows peer IDs for use in other commands (e.g., `channel#123456`)
- Auto-caches peer information for faster resolution

### Reading Messages

#### `tele read <peer>` (aliases: `messages`, `show`, `view`)
Read messages from a conversation.

```bash
tele read @telegram              # Read with sender names (default, 10 per page)
tele read @telegram --page 2     # Jump to page 2
tele read @telegram --limit 20   # 20 messages per page
tele read @telegram --meta       # Show pagination info
tele read @telegram --msg-id     # Show message IDs
tele read @telegram --no-sender  # Hide sender info
tele read @telegram --media-info # Show detailed media info
tele read @telegram --no-more    # Disable pagination
tele read @telegram --since 7d   # Messages from last 7 days
tele read @telegram --since 2024-01-01 --until 2024-01-31
```

**Peer formats:**
- `@username` - Username (e.g., `@telegram`)
- `channel#ID` - Channel ID from `tele chats` output
- `chat#ID` - Chat ID from `tele chats` output
- Numeric user ID (if cached)

**Date filters:**
- `--since`: Relative (`7d`, `1w`, `1m`) or absolute (`YYYY-MM-DD`)
- `--until`: Absolute date (`YYYY-MM-DD`)

### Sending Messages

#### `tele send <peer> [text]` (aliases: `msg`, `text`)
Send a text message or file to a chat.

```bash
tele send @user "Hello!"                    # Send text
tele send @user --file photo.jpg            # Send file (auto-detect type)
tele send @user --file pic.jpg "Check"      # Send file with caption
tele send @user --file image.jpg --as photo # Force send as photo
tele send @user --file video.mp4 --as video # Force send as video
tele send channel#123 "Announcement"        # Send to channel
```

**File types:**
- `--as auto` (default): Auto-detect from mime type
- `--as document`: Send as document
- `--as photo`: Send as photo
- `--as video`: Send as video

**Note:** Caption is currently sent as a separate message (Telegram API limitation).

### Searching Messages

#### `tele search <peer> <query>`
Search for messages containing a specific query.

```bash
tele search @channel "hello"        # Search for "hello" in @channel
tele search @user "keyword" --limit 20
```

### Replying to Messages

#### `tele reply <peer> <msg-id> <text>`
Reply to a specific message.

```bash
tele reply @user 12345 "Thanks!"     # Reply to message 12345
tele reply channel#123 67890 "Got it"
```

Use `tele read --msg-id` to find message IDs.

### Forwarding Messages

#### `tele forward <from-peer> <to-peer> <msg-id>`
Forward a message from one chat to another.

```bash
tele forward @user1 @user2 12345     # Forward msg 12345 from @user1 to @user2
tele forward channel#123 @me 67890   # Forward from channel to saved messages
```

### Deleting Messages

#### `tele delete <peer> <msg-id>`
Delete a message (only your own messages).

```bash
tele delete @user 12345        # Delete message 12345
tele delete channel#123 67890
```

### Downloading Media

#### `tele download <peer>` (aliases: `dl`, `get`)
Download media files (photos, videos, documents) from a chat.

```bash
tele download @channel              # Download all media
tele download @user -d ./media      # Download to ./media directory
tele download @user --msg-id 12345  # Download specific message
tele download @user --index 5        # Download 5th message with media
tele download @user --limit 100     # Scan up to 100 messages
tele download @user --all           # Download all media (no limit)
```

### Peer Information

#### `tele info <peer>`
Show detailed information about a user, chat, or channel.

```bash
tele info @telegram          # Show info about @telegram
tele info channel#123456     # Show info about channel
```

### Version

#### `tele version`
Show version information.

```bash
tele version
```

## Output Formats

### Human-Readable (Default)
- **Table format**: Formatted columns with headers
- **Plain format**: Simple text output

### Machine-Readable
- **TSV format**: Tab-separated values for scripting

Example:
```bash
tele chats --format tsv
tele read @user --format tsv
```

## Pagination

Most commands support pagination:

- **Default**: 10 items per page
- **Interactive**: Press Enter to load more
- **Jump to page**: Use `--page` flag
- **Metadata**: Use `--meta` to show pagination info
- **Disable**: Use `--no-more` to disable interactive prompts

Example:
```bash
tele chats --limit 20 --page 2 --meta
# Output:
# Page 2 (showing 20, more available)
```

## File Storage

All data is stored in `~/.config/tele/`:

- `config.json` - APP_ID and APP_HASH (encrypted)
- `session.json` - Telegram session data
- `peers.json` - Peer cache for faster resolution

## Examples

### Basic Usage

```bash
# Login
tele login

# List conversations
tele chats

# Read messages from a channel
tele read @telegram

# Send a message
tele send @friend "Hello from CLI!"

# Send a photo
tele send @friend --file photo.jpg --as photo "Check this out"

# Download media
tele download @channel -d ./downloads
```

### Advanced Usage

```bash
# Read messages from last week
tele read @channel --since 7d

# Search for specific keyword
tele search @group "meeting"

# Reply to a message
tele reply @user 12345 "Got it, thanks!"

# Forward important message
tele forward @channel @me 67890

# Download specific media
tele read @channel --msg-id
tele download @channel --msg-id 12345
```

### Scripting

```bash
# Get username (stdout)
USERNAME=$(tele login)

# List conversations (TSV format)
tele chats --format tsv --no-more > conversations.tsv

# Read messages (TSV format)
tele read @channel --format tsv --no-more > messages.tsv
```

## Troubleshooting

### "APP_ID not set" Error
Run `tele login` first. Credentials will be saved automatically.

### "FLOOD_WAIT" Error
The client automatically handles rate limits. If you see this message, wait for the specified duration.

### "Peer not found" Error
- Make sure you've run `tele chats` first to populate the peer cache
- Use the exact format from `tele chats` output (e.g., `channel#123456`)
- For usernames, make sure they're correct (case-sensitive)

### File Shows as "unknown"
Make sure you're using `--as photo` or `--as video` for images/videos, or the file has a proper extension.

## Acknowledgments

Built with [gotd/td](https://github.com/gotd/td) - A Go Telegram client library.
