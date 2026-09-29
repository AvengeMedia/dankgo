package lyrics

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCommandSource(t *testing.T) {
	script := filepath.Join(t.TempDir(), "provider")
	body := `#!/bin/sh
case "$(cat)" in
*'"title":"Hit"'*) printf '{"format":"lrc","lyrics":"[00:01.00]hello"}' ;;
*'"title":"Broken"'*) exit 1 ;;
*'"title":"Large"'*) head -c 600000 /dev/zero ;;
esac
`
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	client := New(Options{CacheDir: t.TempDir(), Resolve: func(provider Provider) (Source, bool) {
		return Source{Fetch: Command(script)}, provider == "plugin"
	}})
	lookup := func(title string) (*Result, error) {
		return client.Lookup(context.Background(), Request{Artist: "Artist", Title: title, Providers: []Provider{"plugin"}})
	}

	if result, err := lookup("Hit"); err != nil || result.Source != "plugin" || len(result.Synced) != 1 || result.Synced[0].Text != "hello" {
		t.Fatalf("hit = %+v, %v", result, err)
	}
	if result, err := lookup("Miss"); err != nil || result.Found {
		t.Fatalf("empty output = %+v, %v", result, err)
	}
	if _, err := lookup("Broken"); err == nil {
		t.Fatal("a failing command was reported as missing lyrics")
	}
	if _, err := lookup("Large"); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("oversized output error = %v", err)
	}
}
