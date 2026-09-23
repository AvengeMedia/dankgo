package files

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMimeFromName(t *testing.T) {
	resolver := extensionResolver{}
	assert.Equal(t, "text/markdown", resolver.FromName("readme.MD"))
	assert.Equal(t, "image/png", resolver.FromName("shot.png"))
	assert.Equal(t, "text/plain", resolver.FromName("notes.txt"))
	assert.Empty(t, resolver.FromName("Makefile"))
	assert.Empty(t, resolver.FromName("archive.unknownext"))
}

func TestNeedsSniff(t *testing.T) {
	assert.True(t, needsSniff(""))
	assert.True(t, needsSniff("text/plain"))
	assert.True(t, needsSniff("application/octet-stream"))
	assert.False(t, needsSniff("image/png"))
}

func TestSniffReadsMagic(t *testing.T) {
	root := t.TempDir()

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	picture := filepath.Join(root, "nameless")
	require.NoError(t, os.WriteFile(picture, buf.Bytes(), 0o644))
	assert.Equal(t, "image/png", sniff(picture))

	script := filepath.Join(root, "script")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho hi\n"), 0o755))
	assert.Equal(t, "application/x-shellscript", sniff(script))

	text := filepath.Join(root, "prose")
	require.NoError(t, os.WriteFile(text, []byte("héllo wörld\n"), 0o644))
	assert.Equal(t, "text/plain", sniff(text))

	binary := filepath.Join(root, "blob")
	require.NoError(t, os.WriteFile(binary, []byte{0x00, 0x01, 0x02, 0x7f}, 0o644))
	assert.Equal(t, "application/octet-stream", sniff(binary))

	assert.Empty(t, sniff(filepath.Join(root, "missing")))
}

func TestIconLadders(t *testing.T) {
	assert.Equal(t, []string{"image-png", "image-x-generic", genericIcon}, mimeLadder("image/png", false))
	assert.Equal(t, []string{"application-x-shellscript", "application-x-executable", "application-x-generic", genericIcon}, mimeLadder("application/x-shellscript", true))
	assert.Equal(t, []string{genericIcon}, mimeLadder("", false))
	assert.Equal(t, []string{"application-x-executable", genericIcon}, mimeLadder("", true))
}

func TestFolderIconsFollowUserDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(home, "Documents"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(home, "Elsewhere"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".config", "user-dirs.dirs"), []byte(
		"# comment\nXDG_DOCUMENTS_DIR=\"$HOME/Documents\"\nXDG_MUSIC_DIR=\"$HOME/Tunes\"\n"), 0o644))

	entries, err := List(home, ListOptions{Sort: nameSort()})
	require.NoError(t, err)
	assert.Equal(t, "folder-documents", byName(t, entries, "Documents").IconName)
	assert.Equal(t, "folder", byName(t, entries, "Elsewhere").IconName)
}

func TestWatchEnrichesMimeAfterSnapshot(t *testing.T) {
	bus := &fakeBus{}
	svc := NewService(bus, t.TempDir(), nil)
	t.Cleanup(svc.Close)

	root := t.TempDir()
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4))))
	require.NoError(t, os.WriteFile(filepath.Join(root, "nameless"), buf.Bytes(), 0o644))

	w, err := svc.Watch(context.Background(), root, nameSort(), false)
	require.NoError(t, err)
	svc.Attach([]string{w.topic})

	snapshot, _ := w.page(ListOptions{Sort: nameSort()})
	assert.Empty(t, snapshot.Entries[0].Mime, "the snapshot goes out before sniffing")

	event := bus.await(t, func(b batch) bool { return len(b.Changed) > 0 && b.Changed[0].Name == "nameless" })
	assert.Equal(t, "image/png", event.Changed[0].Mime)
	assert.Equal(t, "image-png", event.Changed[0].IconName)
	assert.True(t, event.Changed[0].Thumbnailable)

	entry, ok := w.entry("nameless")
	require.True(t, ok)
	assert.Equal(t, "image/png", entry.Mime)
}
