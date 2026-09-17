package lyrics

import (
	"context"
	"strings"
	"time"
)

const lrclibURL = "https://lrclib.net/api/get"

type lrclibResponse struct {
	Instrumental bool   `json:"instrumental"`
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
	Lyricsfile   string `json:"lyricsfile"`
}

func (c *Client) fetchLrclib(ctx context.Context, req Request) (*Lyrics, error) {
	query := req.query("track_name", "artist_name", "album_name", "duration")
	var payload lrclibResponse
	if err := c.getJSON(ctx, lrclibURL, query, 8*time.Second, &payload); err != nil {
		return nil, err
	}
	return decodeLrclib(payload)
}

func decodeLrclib(payload lrclibResponse) (*Lyrics, error) {
	if payload.Instrumental {
		return &Lyrics{Instrumental: true}, nil
	}
	var lyricsfileErr error
	if payload.Lyricsfile != "" {
		parsed, err := ParseLyricsfile([]byte(payload.Lyricsfile))
		if err == nil && !parsed.Empty() {
			return parsed, nil
		}
		lyricsfileErr = err
	}

	parsed, err := ParseLRC([]byte(payload.SyncedLyrics))
	if err != nil {
		return nil, err
	}
	if parsed.Plain == "" {
		if strings.Count(payload.PlainLyrics, "\n")+1 > MaxLines {
			return nil, ErrTooLarge
		}
		parsed.Plain = payload.PlainLyrics
	}
	if parsed.Empty() && lyricsfileErr != nil {
		return nil, lyricsfileErr
	}
	return parsed, nil
}
