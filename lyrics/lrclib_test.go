package lyrics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

type lrclibStub struct {
	exact        *lrclibRecord
	records      []lrclibRecord
	searchStatus int
	calls        []string
}

func (s *lrclibStub) RoundTrip(req *http.Request) (*http.Response, error) {
	s.calls = append(s.calls, req.URL.Path)
	var body any = s.records
	status := max(s.searchStatus, http.StatusOK)
	if req.URL.Path == "/api/get" {
		body = s.exact
		if s.exact == nil {
			status = http.StatusNotFound
		}
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(data)), Request: req}, nil
}

func TestLrclibSearchesWhenTheExactSignatureIsPlain(t *testing.T) {
	req := Request{Artist: "Anjan Dutta", Title: "2441139 (Bela Bose)", Album: "YouTube Music", Duration: 312 * time.Second}
	plain := lrclibRecord{TrackName: "2441139 (bela Bose)", ArtistName: "Anjan Dutta, Neel Dutta, Sovan Mukherjee", Duration: 312, PlainLyrics: "Chakrita ami peye gechi"}
	records := []lrclibRecord{
		plain,
		{TrackName: "Bela Bose", ArtistName: "Someone Else", Duration: 311, SyncedLyrics: "[00:22.20]Other song"},
		{TrackName: "2441139 (Bela Bose)", ArtistName: "Anjan Dutta", Duration: 318, SyncedLyrics: "[00:22.20]Farther cut"},
		{TrackName: "2441139 (bela Bose)", ArtistName: "Anjan Dutta; Neel Dutta; Sovan Mukherjee", Duration: 311, SyncedLyrics: "[00:22.20]Closest cut"},
		{TrackName: "2441139 (Bela Bose)", ArtistName: "Anjan Dutt", Duration: 303, SyncedLyrics: "[00:22.20]Short cut"},
	}
	for name, exact := range map[string]*lrclibRecord{"plain": &plain, "missing": nil} {
		t.Run(name, func(t *testing.T) {
			client := isolate(t)
			stub := &lrclibStub{exact: exact, records: records}
			client.http.Transport = stub
			got, err := client.fetchLrclib(context.Background(), req)
			if err != nil || len(got.Synced) != 1 || got.Synced[0].Text != "Closest cut" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if len(stub.calls) != 2 || stub.calls[1] != "/api/search" {
				t.Fatalf("calls = %v", stub.calls)
			}
		})
	}

	t.Run("synced signature skips the search", func(t *testing.T) {
		client := isolate(t)
		stub := &lrclibStub{exact: &records[3], records: records}
		client.http.Transport = stub
		if got, err := client.fetchLrclib(context.Background(), req); err != nil || len(stub.calls) != 1 || got.Synced[0].Text != "Closest cut" {
			t.Fatalf("got %+v, %v, calls %v", got, err, stub.calls)
		}
	})

	t.Run("plain signature outranks plain search results", func(t *testing.T) {
		client := isolate(t)
		stub := &lrclibStub{exact: &plain, records: records[:2]}
		client.http.Transport = stub
		if got, err := client.fetchLrclib(context.Background(), req); err != nil || got.Plain != plain.PlainLyrics {
			t.Fatalf("got %+v, %v", got, err)
		}
	})

	t.Run("failed search is not answered with the plain signature", func(t *testing.T) {
		client := isolate(t)
		client.http.Transport = &lrclibStub{exact: &plain, searchStatus: http.StatusServiceUnavailable}
		var status *StatusError
		if got, err := client.fetchLrclib(context.Background(), req); !errors.As(err, &status) || got != nil {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
}

func TestLrclibBestPrefersWordSyncThenDuration(t *testing.T) {
	req := Request{Artist: "Artist", Title: "Track", Duration: 200 * time.Second}
	records := []lrclibRecord{
		{TrackName: "Track", ArtistName: "Artist", Duration: 200, SyncedLyrics: "[00:01.00]line"},
		{TrackName: "Track", ArtistName: "Artist", Duration: 206, SyncedLyrics: "[00:01.00]line", HasWordSync: true},
		{TrackName: "Track", ArtistName: "Artist", Duration: 204, SyncedLyrics: "[00:01.00]line", HasWordSync: true},
	}
	if best := lrclibBest(records, req); best == nil || best.Duration != 204 {
		t.Fatalf("best = %+v", best)
	}
	req.Duration = 0
	if best := lrclibBest(records, req); best == nil || best.Duration != 206 {
		t.Fatalf("without a duration, best = %+v", best)
	}
}
