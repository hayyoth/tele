package media

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

// Info returns a human-readable description of media in a message.
func Info(msg *tg.Message) string {
	if msg.Media == nil {
		return ""
	}
	switch m := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		if m.Photo != nil {
			if photo, ok := m.Photo.(*tg.Photo); ok && len(photo.Sizes) > 0 {
				if size, ok := photo.Sizes[0].(*tg.PhotoSize); ok {
					return fmt.Sprintf("[photo %dx%d]", size.W, size.H)
				}
			}
		}
		return "[photo]"
	case *tg.MessageMediaDocument:
		if doc, ok := m.Document.(*tg.Document); ok {
			// Check attributes to determine type
			var isVideo, isAudio, isVoice, isSticker, isGif bool
			var duration int
			var title, filename string
			var width, height int

			for _, attr := range doc.Attributes {
				switch a := attr.(type) {
				case *tg.DocumentAttributeFilename:
					filename = a.FileName
				case *tg.DocumentAttributeVideo:
					isVideo = true
					duration = int(a.Duration)
					width = a.W
					height = a.H
				case *tg.DocumentAttributeAudio:
					isAudio = true
					duration = int(a.Duration)
					if a.Title != "" {
						title = a.Title
					}
					if a.Voice {
						isVoice = true
					}
				case *tg.DocumentAttributeSticker:
					isSticker = true
				case *tg.DocumentAttributeAnimated:
					isGif = true
				}
			}

			size := formatSize(doc.Size)
			mime := doc.MimeType
			if mime == "" {
				mime = "file"
			}

			if isVoice {
				return fmt.Sprintf("[voice %s]", formatDuration(duration))
			}
			if isSticker {
				return "[sticker]"
			}
			if isGif {
				return fmt.Sprintf("[gif %s]", size)
			}
			if isVideo {
				if width > 0 && height > 0 {
					return fmt.Sprintf("[video %dx%d %s]", width, height, size)
				}
				return fmt.Sprintf("[video %s]", size)
			}
			if isAudio {
				if title != "" {
					return fmt.Sprintf("[audio %s %s]", title, formatDuration(duration))
				}
				return fmt.Sprintf("[audio %s]", formatDuration(duration))
			}

			// Generic document
			if filename != "" {
				return fmt.Sprintf("[%s %s %s]", mime, filename, size)
			}
			return fmt.Sprintf("[%s %s]", mime, size)
		}
		return "[document]"
	case *tg.MessageMediaContact:
		return "[contact]"
	case *tg.MessageMediaGeo:
		return "[location]"
	case *tg.MessageMediaVenue:
		return "[venue]"
	case *tg.MessageMediaPoll:
		return "[poll]"
	default:
		return fmt.Sprintf("[%T]", m)
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm%ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dh%dm", seconds/3600, (seconds%3600)/60)
}

// HasFile returns true if media has a downloadable file.
func HasFile(msg *tg.Message) bool {
	if msg.Media == nil {
		return false
	}
	switch m := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		return m.Photo != nil
	case *tg.MessageMediaDocument:
		return m.Document != nil
	default:
		return false
	}
}

// FileLocation extracts file location for download.
func FileLocation(msg *tg.Message) (tg.InputFileLocationClass, string, error) {
	if msg.Media == nil {
		return nil, "", fmt.Errorf("no media")
	}
	switch m := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		if photo, ok := m.Photo.(*tg.Photo); ok {
			// Get largest size
			var largest *tg.PhotoSize
			for _, size := range photo.Sizes {
				if s, ok := size.(*tg.PhotoSize); ok {
					if largest == nil || s.W > largest.W {
						largest = s
					}
				}
			}
			if largest != nil {
				return &tg.InputPhotoFileLocation{
					ID:            photo.ID,
					AccessHash:    photo.AccessHash,
					FileReference: photo.FileReference,
					ThumbSize:     largest.Type,
				}, fmt.Sprintf("photo_%d.jpg", photo.ID), nil
			}
		}
		return nil, "", fmt.Errorf("invalid photo")
	case *tg.MessageMediaDocument:
		if doc, ok := m.Document.(*tg.Document); ok {
			name := fmt.Sprintf("file_%d", doc.ID)
			defaultName := name
			for _, attr := range doc.Attributes {
				switch a := attr.(type) {
				case *tg.DocumentAttributeFilename:
					name = a.FileName
				case *tg.DocumentAttributeAudio:
					if a.Voice && name == defaultName {
						name = fmt.Sprintf("voice_%d.ogg", doc.ID)
					}
				case *tg.DocumentAttributeSticker:
					if name == defaultName {
						name = fmt.Sprintf("sticker_%d.webp", doc.ID)
					}
				case *tg.DocumentAttributeAnimated:
					if name == defaultName {
						name = fmt.Sprintf("gif_%d.mp4", doc.ID)
					}
				}
			}
			return &tg.InputDocumentFileLocation{
				ID:            doc.ID,
				AccessHash:    doc.AccessHash,
				FileReference: doc.FileReference,
			}, name, nil
		}
		return nil, "", fmt.Errorf("invalid document")
	default:
		return nil, "", fmt.Errorf("unsupported media type %T", m)
	}
}

// SanitizeFilename removes unsafe characters from filename for cross-platform compatibility.
func SanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	return s
}
