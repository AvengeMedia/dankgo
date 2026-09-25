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
	got := tidyKuGou(parsed)
	want := []Line{{Start: 2, End: 3.5, Text: "Hello world", Words: []Word{{Start: 2, End: 2.5, Text: "Hello "}, {Start: 2.5, End: 3.5, Text: "world"}}}}
	if !reflect.DeepEqual(got.Synced, want) || got.Plain != "Hello world" {
		t.Fatalf("got %+v", got)
	}

	parsed, err = decodeKRC(encodeKRC(t, "[1589,320317]<0,320317,0>纯音乐，请欣赏\n"))
	if err != nil || !tidyKuGou(parsed).Instrumental {
		t.Fatalf("instrumental marker: %+v, %v", parsed, err)
	}
}
