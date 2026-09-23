package files

import (
	"mime"
	"path/filepath"
	"strings"
)

// MimeResolver types an entry. FromName must not touch the disk.
type MimeResolver interface {
	FromName(name string) string
	Sniff(path string) string
}

// IconResolver adds shared-mime-info icons; without it the ladder uses the naming convention.
type IconResolver interface {
	MimeIcon(mimeType string) (specific, generic string)
}

// DesktopResolver reads the Name and Icon of a trusted desktop file.
type DesktopResolver interface {
	DesktopDisplay(path string, trusted bool) (name, icon string, ok bool)
}

type extensionResolver struct{}

var extraTypes = map[string]string{
	"7z":       "application/x-7z-compressed",
	"appimage": "application/x-executable",
	"avif":     "image/avif",
	"conf":     "text/plain",
	"desktop":  "application/x-desktop",
	"go":       "text/x-go",
	"heic":     "image/heif",
	"jxl":      "image/jxl",
	"log":      "text/plain",
	"md":       "text/markdown",
	"mkv":      "video/x-matroska",
	"opus":     "audio/opus",
	"qml":      "text/x-qml",
	"rs":       "text/rust",
	"sh":       "application/x-shellscript",
	"toml":     "application/toml",
	"ts":       "text/x-typescript",
	"webp":     "image/webp",
	"zst":      "application/zstd",
}

func (extensionResolver) FromName(name string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if ext == "" {
		return ""
	}
	if known, ok := extraTypes[ext]; ok {
		return known
	}
	full := mime.TypeByExtension("." + ext)
	if full == "" {
		return ""
	}
	if cut := strings.IndexByte(full, ';'); cut >= 0 {
		full = full[:cut]
	}
	return strings.TrimSpace(full)
}

func (extensionResolver) Sniff(path string) string { return sniff(path) }

func needsSniff(mimeType string) bool {
	switch mimeType {
	case "", "text/plain", "application/octet-stream":
		return true
	default:
		return false
	}
}

// needsDesktop reports trusted desktop files, the only ones listed under their own Name.
func needsDesktop(e Entry) bool { return e.Mime == MimeDesktop && e.Executable }

func mediaType(mimeType string) string {
	cut := strings.IndexByte(mimeType, '/')
	if cut < 0 {
		return ""
	}
	return mimeType[:cut]
}
