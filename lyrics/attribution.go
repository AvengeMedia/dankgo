package lyrics

// Attribution credits the source of a Result. Text is the wording the source asks for, if any.
type Attribution struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Text string `json:"text,omitempty"`
}

var attributions = map[Provider]Attribution{
	BetterLyrics: {Name: "Better Lyrics", URL: "https://betterlyrics.org"},
	Unison:       {Name: "Unison", URL: "https://unison.boidu.dev", Text: "Lyrics from Unison (https://unison.boidu.dev)"},
	LyricsPlus:   {Name: "LyricsPlus", URL: "https://github.com/ibratabian17/lyricsplus"},
	LRCLIB:       {Name: "LRCLIB", URL: "https://lrclib.net"},
}

// Attribution is zero for Sidecar and unknown providers.
func (p Provider) Attribution() Attribution {
	return attributions[p]
}
