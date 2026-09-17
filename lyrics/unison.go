package lyrics

import (
	"context"
	"time"
)

type unisonResponse struct {
	Data struct {
		Lyrics string `json:"lyrics"`
		Format string `json:"format"`
	} `json:"data"`
}

func (c *Client) fetchUnison(ctx context.Context, req Request) (*Lyrics, error) {
	query := req.query("song", "artist", "album", "duration")
	var payload unisonResponse
	if err := c.getJSON(ctx, "https://unison.betterlyrics.org/lyrics", query, 8*time.Second, &payload); err != nil {
		return nil, err
	}
	return decodeUnison(payload)
}

func decodeUnison(payload unisonResponse) (*Lyrics, error) {
	if payload.Data.Format == "ttml" {
		return ParseTTML([]byte(payload.Data.Lyrics))
	}
	return ParseLRC([]byte(payload.Data.Lyrics))
}
