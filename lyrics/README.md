# lyrics

Lyrics lookup for a track. Checks for a local .lrc first, then a disk cache, then LRCLIB / BetterLyrics / Unison / LyricsPlus. Returns word-synced lyrics when a source has them, otherwise line-synced, otherwise plain text.

This is the lyrics engine from DMS, pulled out so other things can use it.

## Usage

```go
client := lyrics.New(lyrics.Options{UserAgent: "myplayer/1.0 (+https://example.com)"})

result, err := client.Lookup(ctx, lyrics.Request{
    Artist: "Daft Punk",
    Title:  "Get Lucky",
})
if err != nil || !result.Found {
    return
}
for _, line := range result.Synced {
    fmt.Println(line.Start.Duration(), line.Text)
}
```

Artist and title are required. The rest of `Request` is optional:

| Field               |                                                                                                                                                 |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `Album`, `Duration` | Passed to the providers to narrow the match. Duration is rounded to the second.                                                                 |
| `FileURL`           | The track's `file://` URL. Turns on the local file lookup below.                                                                                |
| `Providers`         | Priority order. nil = `DefaultProviders()`, which is BetterLyrics, Unison, LyricsPlus, LRCLIB. An empty non-nil slice = no providers, local files only. |
| `CacheOnly`         | Never hit the network. Local files and cached results only.                                                                                     |

Make one `Client` and keep it. The rate limiter and request dedup live on it.

## Demo

```
go run ./lyrics/example "Daft Punk" "Get Lucky"
```

```
source=lyricsplus cached=false lines=114 voices=2

     31.4s  Like the legend of the Phoenix  [6 words]
    35.93s  All ends with beginnings  [4 words]
    39.98s  What keeps the planet spinnin', uh-uh  [7 words]
```

Source is [example/main.go](example/main.go).

## Lookup order

**1. Local files.** Only if `FileURL` is set. For `file:///music/get-lucky.flac` it tries, in the same folder:

```
/music/get-lucky.lyricsfile.yaml
/music/get-lucky.ttml
/music/get-lucky.lrc
/music/get-lucky.LRC
```

That's the whole search, there is no library scan. A local file beats everything else and needs no network. Files over 256 KiB are ignored. MPRIS players report this URL as `xesam:url`.

**2. Disk cache.** `Options.CacheDir`, or `dankgo/lyrics` under `os.UserCacheDir()`. One JSON file per track per provider. Misses are cached too, for 14 days. A line-synced entry older than 14 days is re-requested from its provider once, and replaced only if word sync came back. `Options.DisableUpgrade` turns that off. `client.Prune()` removes entries older than 90 days, it no-ops if it already ran in the last 24h.

**3. Providers.** All requested at the same time, 12 second timeout. Priority still holds: if LRCLIB answers first but BetterLyrics is ahead of it in the list, the LRCLIB result is held until BetterLyrics finishes or the tiemout hits.

| Provider       | Format                    | Sync                                  |
| -------------- | ------------------------- | ------------------------------------- |
| `BetterLyrics` | TTML                      | word, voices, backing vocals          |
| `Unison`       | TTML, LRC or plain        | whatever was submitted                |
| `LyricsPlus`   | JSON                      | word, voices, backing vocals          |
| `LRCLIB`       | LRC, sometimes Lyricsfile | line, word when there is a Lyricsfile |

Each provider gets a burst of 10 requests, then one more every 2 seconds. Over that, `Lookup` returns `ErrRateLimited`.

## Result

```go
type Result struct {
    Found  bool
    Source Provider // which provider, or Sidecar for a local file
    Cached bool
    Lyrics
}

type Lyrics struct {
    Synced       []Line // sorted by Start, empty if the source has no timing
    Plain        string
    Instrumental bool
    Voices       map[VoiceID]Voice
}
```

No lyrics is `Found: false`, not an error. Errors are for when every provider failed, the provider id is unknown, or the context is done.

On a `Line`:

- `Start`, `End` are `Seconds` (a float64) from the start of the track. `.Duration()` converts. `End` is 0 when unknown.
- `Words` is only set for word-synced sources. Concatenating every `Word.Text` gives back `Line.Text` exactly, spaces included.
- `Voice` is a key into `Lyrics.Voices`, which has the singer's name and whether it's a person or a group.
- `Background` is a backing vocal.
- `Group` ties together lines that came from one line in the source, e.g. a lead and the backing vocal under it. 0 = not grouped.

## Parsers

Usable without a `Client`:

```go
lyrics.ParseLRC(data)        // plain, synced, or enhanced (word-synced) LRC
lyrics.ParseTTML(data)
lyrics.ParseLyricsfile(data) // Lyricsfile 1.0 YAML
```

All three are a `lyrics.Parser`, which is just `func([]byte) (*Lyrics, error)`. They expect hostile input and cap it at 512 KiB, 2000 lines, 16000 words, returning `ErrTooLarge` beyond that. An empty document is not an error, check `.Empty()`.

## Adding things

A new file format: write a `Parser`, add a row to `sidecarFormats` in `sidecar.go`.

A new provider: add the `Provider` const, write a `fetchX` method on `Client` that returns `*Lyrics`, add a row to `fetchers` in `providers.go`. Caching, rate limiting and the priority race come for free.

## JSON

`Result` marshals to the format the DMS shell reads. Made-up values, every field shown:

```json
{
  "found": true,
  "source": "lyricsplus",
  "cached": false,
  "synced": [
    {
      "t": 31.4,
      "e": 35.2,
      "x": "Like the legend",
      "w": [
        { "t": 31.4, "e": 31.9, "x": "Like " },
        { "t": 31.9, "e": 32.3, "x": "the " },
        { "t": 32.3, "e": 35.2, "x": "legend" }
      ],
      "voice": "v1"
    },
    {
      "t": 33.0,
      "e": 35.0,
      "x": "ooh",
      "voice": "v1",
      "background": true,
      "group": 4
    }
  ],
  "plain": "Like the legend\nooh",
  "instrumental": false,
  "voices": { "v1": { "name": "Lead singer", "type": "person" } }
}
```

`t` = start, `e` = end, `x` = text, `w` = words. Times are seconds. `e`, `w`, `voice`, `background`, `group`, `voices` are omitted when unset. Cache files use the same shape.
