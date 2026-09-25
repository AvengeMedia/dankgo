package lyrics

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	kugouSongsURL     = "https://mobileservice.kugou.com/api/v3/search/song"
	kugouSearchURL    = "https://lyrics.kugou.com/search"
	kugouDownloadURL  = "https://lyrics.kugou.com/download"
	kugouInstrumental = "纯音乐"
)

var (
	errKRC = errors.New("invalid krc document")

	krcKey  = []byte{64, 71, 97, 119, 94, 50, 116, 71, 81, 54, 49, 45, 206, 210, 110, 105}
	krcLine = regexp.MustCompile(`^\[(\d+),(\d+)\](.*)`)
	krcWord = regexp.MustCompile(`<(\d+),(\d+),\d+>([^<]*)`)
)

type kugouSong struct {
	Name     string `json:"songname"`
	Singers  string `json:"singername"`
	Hash     string `json:"hash"`
	Duration int    `json:"duration"`
}

type kugouCandidate struct {
	ID        string `json:"id"`
	AccessKey string `json:"accesskey"`
	Song      string `json:"song"`
	Singers   string `json:"singer"`
}

// KRC carries word timing, LRC only lines.
func (c *Client) fetchKuGou(ctx context.Context, req Request) (*Lyrics, error) {
	candidate, err := c.kugouCandidate(ctx, req)
	if err != nil {
		return nil, err
	}
	if found, err := c.kugouDownload(ctx, candidate, "krc", decodeKRC); err == nil && !found.Empty() {
		return found, nil
	}
	return c.kugouDownload(ctx, candidate, "lrc", ParseLRC)
}

func (c *Client) kugouCandidate(ctx context.Context, req Request) (kugouCandidate, error) {
	keyword := req.Title + " " + req.Artist
	query := url.Values{"version": {"9108"}, "plat": {"0"}, "pagesize": {"8"}, "showtype": {"0"}, "keyword": {keyword}}
	var songs struct {
		Data struct {
			Info []kugouSong `json:"info"`
		} `json:"data"`
	}
	if err := c.getJSON(ctx, kugouSongsURL, query, 4*time.Second, &songs); err != nil {
		return kugouCandidate{}, err
	}
	for _, song := range songs.Data.Info {
		if !req.matchesDuration(song.Duration) || !kugouMatches(req, song.Name, song.Singers) {
			continue
		}
		candidate, err := c.kugouSearch(ctx, req, url.Values{"hash": {song.Hash}})
		if !errors.Is(err, errNotFound) {
			return candidate, err
		}
	}

	query = url.Values{"keyword": {keyword}}
	if seconds := req.seconds(); seconds > 0 {
		query.Set("duration", strconv.Itoa(seconds*1000))
	}
	return c.kugouSearch(ctx, req, query)
}

func (c *Client) kugouSearch(ctx context.Context, req Request, query url.Values) (kugouCandidate, error) {
	query.Set("ver", "1")
	query.Set("man", "yes")
	query.Set("client", "pc")
	var payload struct {
		Candidates []kugouCandidate `json:"candidates"`
	}
	if err := c.getJSON(ctx, kugouSearchURL, query, 4*time.Second, &payload); err != nil {
		return kugouCandidate{}, err
	}
	for _, candidate := range payload.Candidates {
		if kugouMatches(req, candidate.Song, candidate.Singers) {
			return candidate, nil
		}
	}
	return kugouCandidate{}, errNotFound
}

func kugouMatches(req Request, title, singers string) bool {
	return req.matchesTrack(title, strings.Split(singers, "、"))
}

func (c *Client) kugouDownload(ctx context.Context, candidate kugouCandidate, format string, parse Parser) (*Lyrics, error) {
	query := url.Values{"ver": {"1"}, "client": {"pc"}, "charset": {"utf8"}, "fmt": {format}, "id": {candidate.ID}, "accesskey": {candidate.AccessKey}}
	var payload struct {
		Content string `json:"content"`
	}
	if err := c.getJSON(ctx, kugouDownloadURL, query, 4*time.Second, &payload); err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(payload.Content)
	if err != nil {
		return nil, err
	}
	parsed, err := parse(data)
	if err != nil {
		return nil, err
	}
	return tidyKuGou(parsed), nil
}

// decodeKRC reads KuGou's karaoke lyrics: "krc1", then zlib data XORed with a fixed key.
func decodeKRC(data []byte) (*Lyrics, error) {
	body, ok := bytes.CutPrefix(data, []byte("krc1"))
	if !ok {
		return nil, errKRC
	}
	compressed := make([]byte, len(body))
	for i, value := range body {
		compressed[i] = value ^ krcKey[i%len(krcKey)]
	}
	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	text, err := readBounded(reader)
	if err != nil {
		return nil, err
	}
	return parseKRC(text)
}

func parseKRC(data []byte) (*Lyrics, error) {
	text, err := decode(data)
	if err != nil {
		return nil, err
	}
	var (
		lines []Line
		plain []string
		words int
	)
	for raw := range strings.Lines(text) {
		match := krcLine.FindStringSubmatch(strings.TrimRight(raw, "\r\n"))
		if match == nil {
			continue
		}
		start, end, ok := krcSpan(match[1], match[2])
		if !ok {
			continue
		}
		line := Line{Start: start, End: end}
		for _, word := range krcWord.FindAllStringSubmatch(match[3], -1) {
			from, to, ok := krcSpan(word[1], word[2])
			if !ok {
				continue
			}
			line.Words = append(line.Words, Word{Start: start + from, End: start + to, Text: word[3]})
		}
		line.Words = trimWords(line.Words)
		line.Text = joinWords(line.Words)
		if line.Text == "" {
			continue
		}
		words += len(line.Words)
		if len(lines) >= MaxLines || words > MaxWords {
			return nil, ErrTooLarge
		}
		lines = append(lines, line)
		plain = append(plain, line.Text)
	}
	sortLines(lines)
	return &Lyrics{Synced: lines, Plain: strings.Join(plain, "\n")}, nil
}

func krcSpan(offset, length string) (Seconds, Seconds, bool) {
	start, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		return 0, 0, false
	}
	span, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return milliseconds(start), milliseconds(start) + milliseconds(span), true
}

// tidyKuGou drops the "title - artist" line and "role：name" credits KuGou puts before the lyrics.
func tidyKuGou(parsed *Lyrics) *Lyrics {
	lines := parsed.Synced
	if len(lines) == 0 {
		return parsed
	}
	if len(lines) == 1 && strings.Contains(lines[0].Text, kugouInstrumental) {
		return &Lyrics{Instrumental: true}
	}
	credits := 0
	for i := 1; i < len(lines) && strings.ContainsAny(lines[i].Text, ":："); i++ {
		credits = i + 1
	}
	lyrics := lines[credits:]
	plain := make([]string, len(lyrics))
	for i, line := range lyrics {
		plain[i] = line.Text
	}
	return &Lyrics{Synced: lyrics, Plain: strings.Join(plain, "\n")}
}
