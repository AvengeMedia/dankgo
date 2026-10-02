# lyrics

Lyrics lookup for a track. Checks for a local .lrc first, then a disk cache, then LRCLIB / BetterLyrics / Unison / LyricsPlus / KuGou / YouTube Music. Returns word-synced lyrics when a source has them, otherwise line-synced, otherwise plain text.

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
| `Providers`         | Priority order. nil = `DefaultProviders()`, which is BetterLyrics, Unison, LyricsPlus, KuGou, LRCLIB, YouTubeMusic. An empty non-nil slice = no providers, local files only. |
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

**2. Disk cache.** `Options.CacheDir`, or `dankgo/lyrics` under `os.UserCacheDir()`. One JSON file per track per provider. Misses are cached too, for 14 days. Not a BetterLyrics 401 though, that only means their server has not cached the track yet. A line-synced entry older than 14 days is re-requested from its provider once, and replaced only if word sync came back. `Options.DisableUpgrade` turns that off. `client.Prune()` removes entries older than 90 days, it no-ops if it already ran in the last 24h.

**3. Providers.** All requested at the same time, 12 second timeout. Priority still holds: if LRCLIB answers first but BetterLyrics is ahead of it in the list, the LRCLIB result is held until BetterLyrics finishes or the tiemout hits.

Word sync outranks priority. A word-synced result wins as soon as every word-capable provider ahead of it has finished. When the priority winner is only line-synced or plain and a word-capable provider is still out, it is held for up to 2 seconds in case word sync shows up. LRCLIB is never waited on for this, it has word sync too rarely.

A synced result whose every stamp is zero counts as a miss. LyricsPlus returns those for some Apple tracks, and the next provider usually has real timing.

| Provider       | Format                    | Sync                                  |
| -------------- | ------------------------- | ------------------------------------- |
| `BetterLyrics` | TTML                      | word, voices, backing vocals          |
| `Unison`       | TTML, LRC or plain        | whatever was submitted                |
| `LyricsPlus`   | JSON                      | word, voices, backing vocals          |
| `KuGou`        | KRC, LRC if KRC fails     | word, line from the LRC fallback      |
| `LRCLIB`       | LRC, sometimes Lyricsfile | line, word when there is a Lyricsfile |
| `YouTubeMusic` | JSON                      | line, plain when a line has no timing |

Each provider gets a burst of 10 requests, then one more every 2 seconds. Over that, `Lookup` returns `ErrRateLimited`.

## Result

```go
type Result struct {
    Found  bool
    Source Provider // which provider, or Sidecar for a local file
    Cached bool
    Attribution Attribution // Name, URL, and Text when the source asks for specific wording
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

To write lyrics out again, `result.AsLyricsfile()` encodes any `*Lyrics` (from any parser or provider) back into Lyricsfile 1.0 YAML, and hands back the original bytes untouched when the source was already Lyricsfile. The format requires `title`/`artist`, so other sources need `Lyrics.Title`/`Lyrics.Artist` set: parsers fill what their source carries, callers fill the rest. `Voices`, `Background` and `Group` are dropped, the format has no fields for them.

## Adding things

A new file format: write a `Parser`, add a row to `sidecarFormats` in `sidecar.go`.

A new built-in provider: add the `Provider` const, write a `fetchX` method on `Client` that returns `*Lyrics`, add a row to `builtins` in `providers.go`. Caching, rate limiting and the priority race come for free.

## Custom providers

A provider is a `Source`: an `Attribution`, a `WordSync` flag and a `Fetch` func. The built-in ones are Sources too. `Options.Resolve` supplies Sources for ids that are not built in. It runs on every `Lookup`, so the set can change while the client lives.

```go
client := lyrics.New(lyrics.Options{
    Resolve: func(id lyrics.Provider) (lyrics.Source, bool) {
        if id != "myprovider" {
            return lyrics.Source{}, false
        }
        return lyrics.Source{
            Attribution: lyrics.Attribution{Name: "My Provider", URL: "https://example.com"},
            Fetch:       lyrics.Command("/usr/lib/myprovider/fetch"),
        }, true
    },
})

result, err := client.Lookup(ctx, lyrics.Request{
    Artist:    "Daft Punk",
    Title:     "Get Lucky",
    Providers: []lyrics.Provider{"myprovider", lyrics.LRCLIB},
})
```

A custom Source gets the same cache, rate limit and priority race as a built-in one. It only runs when it is in `Providers`, `DefaultProviders()` is built-ins only. Set `WordSync` only when the source usually has word timing, it lets the source hold a line-synced winner for up to 2 seconds.

`Fetch` returns empty `Lyrics` for no lyrics, which is cached for 14 days like any miss. Errors are not cached.

### Command

`lyrics.Command(name, args...)` is a `Fetch` that runs a program per lookup, so a provider can be written in anything. The track comes in on stdin. `album` and `duration` are left out when unknown:

```json
{ "title": "Get Lucky", "artist": "Daft Punk", "album": "Random Access Memories", "duration": 369 }
```

The answer goes to stdout:

```json
{ "format": "ttml", "lyrics": "<tt xmlns=\"http://www.w3.org/ns/ttml\">...</tt>" }
```

`format` is `lrc`, `ttml` or `lyricsfile`, read with the parsers above. Plain text goes as `lrc`. `{"instrumental": true}` is an instrumental. No output is no lyrics, a non-zero exit is an error. stderr is dropped.

Output over 512 KiB is `ErrTooLarge`. After 8 seconds, or once the lookup has a winner, the command is killed along with everything it started.

Try one by hand:

```
echo '{"title":"Get Lucky","artist":"Daft Punk"}' | ./fetch
```

## JSON

`Result` marshals to the format the DMS shell reads. Made-up values, every field shown:

```json
{
  "found": true,
  "source": "lyricsplus",
  "cached": false,
  "attribution": { "name": "LyricsPlus", "url": "https://github.com/ibratabian17/lyricsplus" },
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
