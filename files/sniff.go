package files

import (
	"bytes"
	"os"
	"unicode/utf8"
)

// sniffExtent covers the longest signature below.
const sniffExtent = 512

type signature struct {
	offset int
	magic  []byte
	mime   string
}

var signatures = []signature{
	{0, []byte("\x89PNG\r\n\x1a\n"), "image/png"},
	{0, []byte("\xff\xd8\xff"), "image/jpeg"},
	{0, []byte("GIF87a"), "image/gif"},
	{0, []byte("GIF89a"), "image/gif"},
	{0, []byte("BM"), "image/bmp"},
	{0, []byte("II*\x00"), "image/tiff"},
	{0, []byte("MM\x00*"), "image/tiff"},
	{0, []byte("%PDF-"), "application/pdf"},
	{0, []byte("\x7fELF"), "application/x-executable"},
	{0, []byte("PK\x03\x04"), "application/zip"},
	{0, []byte("\x1f\x8b"), "application/gzip"},
	{0, []byte("\x28\xb5\x2f\xfd"), "application/zstd"},
	{0, []byte("7z\xbc\xaf\x27\x1c"), "application/x-7z-compressed"},
	{0, []byte("ustar"), "application/x-tar"},
	{0, []byte("OggS"), "audio/ogg"},
	{0, []byte("fLaC"), "audio/flac"},
	{0, []byte("ID3"), "audio/mpeg"},
	{0, []byte("#!"), "application/x-shellscript"},
	{0, []byte("<?xml"), "application/xml"},
	{0, []byte("\x1aE\xdf\xa3"), "video/x-matroska"},
	{4, []byte("ftyp"), "video/mp4"},
	{8, []byte("WEBP"), "image/webp"},
	{8, []byte("AVI "), "video/x-msvideo"},
	{8, []byte("WAVE"), "audio/x-wav"},
}

func sniff(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	head := make([]byte, sniffExtent)
	n, _ := f.Read(head)
	if n == 0 {
		return ""
	}
	head = head[:n]

	for _, sig := range signatures {
		end := sig.offset + len(sig.magic)
		if end > len(head) {
			continue
		}
		if bytes.Equal(head[sig.offset:end], sig.magic) {
			return sig.mime
		}
	}
	if looksTextual(head) {
		return "text/plain"
	}
	return "application/octet-stream"
}

func looksTextual(head []byte) bool {
	if !utf8.Valid(head) && !utf8.Valid(trimPartialRune(head)) {
		return false
	}
	for _, b := range head {
		if b == 0 {
			return false
		}
		if b < 0x09 || (b > 0x0d && b < 0x20 && b != 0x1b) {
			return false
		}
	}
	return true
}

func trimPartialRune(head []byte) []byte {
	for i := len(head) - 1; i >= 0 && i > len(head)-5; i-- {
		if utf8.RuneStart(head[i]) {
			return head[:i]
		}
	}
	return head
}
