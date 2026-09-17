// go run ./lyrics/example "Daft Punk" "Get Lucky"
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/AvengeMedia/dankgo/lyrics"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: example <artist> <title> [album]")
		os.Exit(2)
	}
	req := lyrics.Request{
		Artist:    os.Args[1],
		Title:     os.Args[2],
		Providers: []lyrics.Provider{lyrics.Unison},
	}
	if len(os.Args) > 3 {
		req.Album = os.Args[3]
	}

	client := lyrics.New(lyrics.Options{UserAgent: "dankgo-lyrics-example (+https://github.com/AvengeMedia/dankgo)"})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.Lookup(ctx, req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lookup failed:", err)
		os.Exit(1)
	}
	if !result.Found {
		fmt.Println("no lyrics")
		return
	}

	fmt.Printf("source=%s cached=%t lines=%d voices=%d\n\n", result.Source, result.Cached, len(result.Synced), len(result.Voices))
	if result.Instrumental {
		fmt.Println("(instrumental)")
		return
	}
	if len(result.Synced) == 0 {
		fmt.Println(result.Plain)
		return
	}
	for _, line := range result.Synced {
		fmt.Println(format(line, result.Voices))
	}
}

func format(line lyrics.Line, voices map[lyrics.VoiceID]lyrics.Voice) string {
	stamp := line.Start.Duration().Truncate(10 * time.Millisecond)
	text := line.Text
	if line.Background {
		text = "(" + text + ")"
	}
	if name := voices[line.Voice].Name; name != "" {
		text = name + ": " + text
	}
	if len(line.Words) > 0 {
		text += fmt.Sprintf("  [%d words]", len(line.Words))
	}
	return fmt.Sprintf("%10s  %s", stamp, text)
}
