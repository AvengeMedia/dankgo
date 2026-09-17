package lyrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	rateBurst  = 10
	rateRefill = 0.5
	wordGrace  = 2 * time.Second
)

var ErrRateLimited = errors.New("rate limited")

type rateBucket struct {
	tokens float64
	last   time.Time
}

type Options struct {
	// CacheDir defaults to dankgo/lyrics under os.UserCacheDir.
	CacheDir  string
	UserAgent string
	// DisableUpgrade stops cached line-synced lyrics from being re-requested for word sync.
	DisableUpgrade bool
}

// Client is safe for concurrent use. Rate limits and request dedup are per Client, so share one.
type Client struct {
	cacheDir  string
	userAgent string
	upgrade   bool
	wordGrace time.Duration
	http      *http.Client
	fetch     func(context.Context, Provider, Request) (*Lyrics, error)
	lookups   singleflight.Group

	rateMu      sync.Mutex
	rateBuckets map[Provider]*rateBucket
}

func New(opts Options) *Client {
	client := &Client{
		cacheDir:    opts.CacheDir,
		userAgent:   opts.UserAgent,
		upgrade:     !opts.DisableUpgrade,
		wordGrace:   wordGrace,
		http:        newHTTPClient(),
		rateBuckets: map[Provider]*rateBucket{},
	}
	client.fetch = client.fetchProvider
	if client.userAgent == "" {
		client.userAgent = defaultUserAgent
	}
	if client.cacheDir != "" {
		return client
	}
	if dir, err := os.UserCacheDir(); err == nil {
		client.cacheDir = filepath.Join(dir, "dankgo", "lyrics")
	}
	return client
}

// Lookup checks sidecar files, then the cache, then the providers.
// No lyrics is Found false, not an error.
func (c *Client) Lookup(ctx context.Context, req Request) (*Result, error) {
	if result, ok := fromSidecar(req.FileURL); ok {
		return result, nil
	}
	if req.Artist == "" || req.Title == "" {
		return &Result{}, nil
	}

	providers := req.Providers
	if providers == nil {
		providers = DefaultProviders()
	}
	seen := make(map[Provider]bool, len(providers))
	ordered := make([]Provider, 0, len(providers))
	names := make([]string, 0, len(providers))
	for _, provider := range providers {
		if seen[provider] {
			continue
		}
		if _, known := fetchers[provider]; !known {
			return nil, fmt.Errorf("unknown lyrics provider: %s", provider)
		}
		seen[provider] = true
		ordered = append(ordered, provider)
		names = append(names, string(provider))
	}
	if len(ordered) == 0 {
		return &Result{}, nil
	}
	key := fmt.Sprintf("%s/%t/%s", cacheKey(req), req.CacheOnly, strings.Join(names, ","))
	flight := c.lookups.DoChan(key, func() (any, error) {
		return c.lookupProviders(context.WithoutCancel(ctx), ordered, req)
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case response := <-flight:
		if response.Err != nil {
			return nil, response.Err
		}
		return response.Val.(*Result), nil
	}
}

type providerResponse struct {
	index  int
	result *Result
	err    error
}

// lookupProviders races all providers concurrently while selecting results in
// caller-specified priority order. Fast lower-priority results are buffered
// until every provider ahead of them has failed or missed (with timeout considered)
//
// Word sync outranks priority: a word-synced result wins once every wordProvider ahead
// of it has finished, and a lesser winner is held up to wordGrace for the wordProviders
// still out.
func (c *Client) lookupProviders(ctx context.Context, providers []Provider, req Request) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	replies := make(chan providerResponse, len(providers))
	// Start every enabled provider lookup simultaneously, irregardless of priority
	for index, provider := range providers {
		go func() {
			result, err := c.lookupProvider(ctx, provider, req)
			replies <- providerResponse{index: index, result: result, err: err}
		}()
	}

	// Buffer out-of-order responses until all providers ahead (higher prio) have finished
	finished := make([]bool, len(providers))
	results := make([]providerResponse, len(providers))

	// next is the highest-prio provider whose result hasnt been received yet
	next := 0
	answered := false
	var lookupErr error
	var grace <-chan time.Time
	for next < len(providers) {
		select {
		case <-ctx.Done():
			// We can get lower-priority results faster than the higher-priority provider results
			// we wait for the higher priority until the timeout, then return the best one we already got
			if best := bestFound(results); best != nil {
				return best, nil
			}
			return nil, ctx.Err()
		case <-grace:
			return bestFound(results), nil
		case response := <-replies:
			finished[response.index] = true
			results[response.index] = response
		}
		if word := wordWinner(providers, finished, results); word != nil {
			return word, nil
		}
		for next < len(providers) && finished[next] {
			response := results[next]

			// First success in priority order wins, unless word sync may still arrive
			if response.err == nil && response.result != nil && response.result.Found {
				if !wordPending(providers, finished) {
					return response.result, nil
				}
				if grace == nil {
					grace = time.After(c.wordGrace)
				}
				break
			}

			if response.err != nil {
				lookupErr = response.err
			} else {
				answered = true
			}

			next++
		}
	}
	// A provider that says "no lyrics" outranks one that merely failed to answer.
	if lookupErr != nil && !answered {
		return nil, lookupErr
	}
	return &Result{}, nil
}

// wordWinner is the first word-synced result with no wordProvider still out ahead of it.
func wordWinner(providers []Provider, finished []bool, results []providerResponse) *Result {
	for index, provider := range providers {
		if !finished[index] && wordProviders[provider] {
			return nil
		}
		if result := results[index].result; result != nil && result.Found && result.wordSynced() {
			return result
		}
	}
	return nil
}

func wordPending(providers []Provider, finished []bool) bool {
	for index, provider := range providers {
		if !finished[index] && wordProviders[provider] {
			return true
		}
	}
	return false
}

// bestFound prefers word sync, priority order breaks ties.
func bestFound(responses []providerResponse) *Result {
	var best *Result
	for _, response := range responses {
		if response.result == nil || !response.result.Found {
			continue
		}
		if response.result.wordSynced() {
			return response.result
		}
		if best == nil {
			best = response.result
		}
	}
	return best
}

func (c *Client) lookupProvider(ctx context.Context, provider Provider, req Request) (*Result, error) {
	key := providerCacheKey(provider, req)
	if entry, ok := c.loadCache(key); ok {
		if c.upgrade && !req.CacheOnly && entry.awaitsWordSync() {
			c.upgradeCache(ctx, key, provider, req, entry)
		}
		return fromCache(entry), nil
	}
	if req.CacheOnly {
		return &Result{}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !c.allowRequest(provider) {
		return nil, ErrRateLimited
	}
	found, err := c.fetch(ctx, provider, req)
	if errors.Is(err, errUncached) {
		return &Result{}, nil
	}
	if err != nil && !errors.Is(err, errNotFound) {
		return nil, err
	}
	if err != nil || found.Empty() {
		c.storeCache(key, &cacheEntry{Miss: true})
		return &Result{}, nil
	}
	c.storeCache(key, &cacheEntry{Source: provider, Lyrics: *found})
	return &Result{Found: true, Source: provider, Attribution: provider.Attribution(), Lyrics: *found}, nil
}

// upgradeCache swaps a line-synced entry for word sync once the provider has it.
// The entry is restamped either way, so the provider is asked once per upgradeTTL.
func (c *Client) upgradeCache(ctx context.Context, key string, provider Provider, req Request, entry *cacheEntry) {
	if !c.allowRequest(provider) {
		return
	}
	found, err := c.fetch(ctx, provider, req)
	if err != nil && !errors.Is(err, errNotFound) {
		return
	}
	if err == nil && found.wordSynced() {
		entry.Lyrics = *found
	}
	c.storeCache(key, entry)
}

func fromCache(entry *cacheEntry) *Result {
	if entry.Miss {
		return &Result{Cached: true}
	}
	return &Result{Found: true, Source: entry.Source, Cached: true, Attribution: entry.Source.Attribution(), Lyrics: entry.Lyrics}
}

func (c *Client) allowRequest(provider Provider) bool {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	now := time.Now()
	bucket := c.rateBuckets[provider]
	if bucket == nil {
		bucket = &rateBucket{tokens: rateBurst, last: now}
		c.rateBuckets[provider] = bucket
	}
	bucket.tokens = min(rateBurst, bucket.tokens+now.Sub(bucket.last).Seconds()*rateRefill)
	bucket.last = now

	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens--
	return true
}
