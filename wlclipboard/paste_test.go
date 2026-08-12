//go:build unix

package wlclipboard

import "testing"

func TestExpandOffers(t *testing.T) {
	text := ExpandOffers([]byte("hi"), "text/plain;charset=utf-8")
	if len(text) != 5 {
		t.Fatalf("expected 5 text offers, got %d", len(text))
	}
	seen := map[string]bool{}
	for _, o := range text {
		if string(o.Data) != "hi" {
			t.Errorf("offer %s has wrong data", o.MimeType)
		}
		if seen[o.MimeType] {
			t.Errorf("duplicate offer %s", o.MimeType)
		}
		seen[o.MimeType] = true
	}
	if !seen["UTF8_STRING"] || !seen["STRING"] || !seen["TEXT"] || !seen["text/plain"] {
		t.Errorf("missing X11 alias offers: %v", seen)
	}

	img := ExpandOffers([]byte{1}, "image/png")
	if len(img) != 1 || img[0].MimeType != "image/png" {
		t.Errorf("non-text mime should not expand, got %+v", img)
	}
}

func TestTextOffers(t *testing.T) {
	offers := TextOffers("hello")
	if len(offers) != 5 {
		t.Fatalf("expected 5 offers, got %d", len(offers))
	}
	if offers[0].MimeType != "text/plain;charset=utf-8" {
		t.Errorf("primary offer = %s, want text/plain;charset=utf-8", offers[0].MimeType)
	}
	for _, o := range offers {
		if string(o.Data) != "hello" {
			t.Errorf("offer %s has wrong data", o.MimeType)
		}
	}
}

func TestSelectPreferredMimeType(t *testing.T) {
	tests := []struct {
		name  string
		mimes []string
		want  string
	}{
		{"empty", nil, ""},
		{"utf8 preferred over plain", []string{"image/png", "text/plain", "text/plain;charset=utf-8"}, "text/plain;charset=utf-8"},
		{"plain over legacy", []string{"STRING", "text/plain"}, "text/plain"},
		{"image when no text", []string{"image/jpeg", "image/png"}, "image/png"},
		{"unknown falls back to first", []string{"application/x-foo", "application/x-bar"}, "application/x-foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectPreferredMimeType(tt.mimes); got != tt.want {
				t.Errorf("selectPreferredMimeType(%v) = %q, want %q", tt.mimes, got, tt.want)
			}
		})
	}
}
