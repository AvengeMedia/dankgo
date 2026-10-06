package lyrics

import (
	"bytes"
	"compress/zlib"
	"reflect"
	"testing"
)

func encodeKRC(t *testing.T, text string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	writer.Write([]byte(text))
	writer.Close()
	data := compressed.Bytes()
	for i := range data {
		data[i] ^= krcKey[i%len(krcKey)]
	}
	return append([]byte("krc1"), data...)
}

func TestKuGouKRC(t *testing.T) {
	parsed, err := decodeKRC(encodeKRC(t, "[ti:Song]\n"+
		"[0,1000]<0,500,0>Song<500,500,0> - Artist\n"+
		"[1000,1000]<0,1000,0>Lyrics by：Writer\n"+
		"[2000,1500]<0,500,0>Hello <500,1000,0>world\n"))
	if err != nil {
		t.Fatal(err)
	}
	got := tidyKuGou(Request{}, parsed)
	want := []Line{{Start: 2, End: 3.5, Text: "Hello world", Words: []Word{{Start: 2, End: 2.5, Text: "Hello "}, {Start: 2.5, End: 3.5, Text: "world"}}}}
	if !reflect.DeepEqual(got.Synced, want) || got.Plain != "Hello world" {
		t.Fatalf("got %+v", got)
	}

	parsed, err = decodeKRC(encodeKRC(t, "[1589,320317]<0,320317,0>纯音乐，请欣赏\n"))
	if err != nil || !tidyKuGou(Request{}, parsed).Instrumental {
		t.Fatalf("instrumental marker: %+v, %v", parsed, err)
	}
}

func TestKuGouHeaderShapes(t *testing.T) {
	for _, header := range []string{
		"",
		"[0,80]<0,80,0>作曲 : glass beach\n",
		"[0,34540]<0,34540,0>Smells Like Teen Spirit - Nirvana (涅槃)\n",
		"[0,1370]<0,1370,0>Anti-Hero - Taylor Swift\n[1370,1370]<0,1370,0>Lyrics by：Jack Antonoff\n",
		"[0,1167]<0,260,0>Alexandria <260,260,0>-- <520,260,0>The <780,387,0>Fool\n",
	} {
		parsed, err := decodeKRC(encodeKRC(t, header+"[5000,1000]<0,1000,0>Na na na\n"))
		if err != nil {
			t.Fatal(err)
		}
		if got := tidyKuGou(Request{}, parsed); got.Plain != "Na na na" {
			t.Errorf("%q: %+v", header, got.Synced)
		}
	}
}

func TestKuGouDropsBakedTranslations(t *testing.T) {
	english := func(req Request, text string) string {
		parsed, err := decodeKRC(encodeKRC(t, text))
		if err != nil {
			t.Fatal(err)
		}
		return tidyKuGou(req, parsed).Plain
	}
	roman := Request{Title: "The Fool", Artist: "Alexandria"}
	alternating := "[0,1000]<0,1000,0>The fool jumps on in\n[1000,500]<0,500,0>愚者纵身跃入\n[2000,1000]<0,1000,0>They say he's a madman\n[3000,500]<0,500,0>世人皆言其疯狂\n"
	if got := english(roman, alternating); got != "The fool jumps on in\nThey say he's a madman" {
		t.Errorf("alternating translation kept: %q", got)
	}
	mixed := "[0,1000]<0,1000,0>Hello baby\n[1000,500]<0,500,0>사랑해\n[2000,500]<0,500,0>영원히\n[3000,1000]<0,1000,0>Hello baby\n"
	if got := english(roman, mixed); got != "Hello baby\n사랑해\n영원히\nHello baby" {
		t.Errorf("mixed-language lyrics lost lines: %q", got)
	}
	cjkSong := "[0,1000]<0,1000,0>夢ならば\n[1000,500]<0,500,0>どれほど\n"
	if got := english(Request{Title: "Lemon", Artist: "Kenshi Yonezu"}, cjkSong); got != "夢ならば\nどれほど" {
		t.Errorf("romanized request gutted a CJK song: %q", got)
	}
	romaji := "[0,1000]<0,1000,0>yume naraba\n[1000,500]<0,500,0>夢ならば\n[2000,1000]<0,1000,0>dorehodo\n[3000,500]<0,500,0>どれほど\n"
	if got := english(Request{Title: "Lemon", Artist: "Kenshi Yonezu"}, romaji); got != "yume naraba\n夢ならば\ndorehodo\nどれほど" {
		t.Errorf("romaji upload lost the lyrics: %q", got)
	}
	hangul := "[0,1000]<0,1000,0>saranghae\n[1000,500]<0,500,0>사랑해\n[2000,1000]<0,1000,0>yeongwonhi\n[3000,500]<0,500,0>영원히\n"
	if got := english(roman, hangul); got != "saranghae\n사랑해\nyeongwonhi\n영원히" {
		t.Errorf("romanized korean upload lost the lyrics: %q", got)
	}
}
