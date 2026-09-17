package lyrics

import (
	"context"
	"time"
)

func (c *Client) fetchBetterLyrics(ctx context.Context, req Request) (*Lyrics, error) {
	query := req.query("s", "a", "al", "d")
	var payload struct {
		TTML string `json:"ttml"`
	}
	if err := c.getJSON(ctx, "https://lyrics-api.boidu.dev/getLyrics", query, 4*time.Second, &payload); err != nil {
		return nil, err
	}
	if payload.TTML == "" {
		return nil, errNotFound
	}
	return ParseTTML([]byte(payload.TTML))
}
