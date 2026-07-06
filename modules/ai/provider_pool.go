package ai

import (
	"strings"
	"sync"

	"github.com/masudur-rahman/khorcha-pati/configs"
)

// defaultStickyWindow is how many requests a provider serves before the pool rotates to the
// next one. Rotating in blocks (rather than per request) keeps bursts under each free tier's
// per-minute cap while still spreading load across both providers.
const defaultStickyWindow = 15

// rotationStore persists the pool cursor: the active provider index and how many requests it
// has served in the current window. memStore keeps it in process memory; a Redis-backed store
// can implement the same interface later for multi-instance deployments.
type rotationStore interface {
	load() (idx, used int)
	store(idx, used int)
}

// memStore is the in-memory rotationStore used by default.
type memStore struct {
	mu   sync.Mutex
	idx  int
	used int
}

func (m *memStore) load() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.idx, m.used
}

func (m *memStore) store(idx, used int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idx, m.used = idx, used
}

// providerPool serves AI classification across the configured providers (Gemini first, then
// OpenRouter). It sticks to one provider for a window of requests, then rotates; a rate-limited
// request fails over to the next provider immediately via the ordered sequence().
type providerPool struct {
	mu     sync.Mutex
	order  []Classifier
	window int
	state  rotationStore
}

var (
	poolOnce    sync.Once
	defaultPool *providerPool
)

// getPool returns the process-wide provider pool, built once from the loaded config.
func getPool() *providerPool {
	poolOnce.Do(func() {
		window := configs.TrackerConfig.System.AIStickyWindow
		if window <= 0 {
			window = defaultStickyWindow
		}
		defaultPool = &providerPool{
			order:  buildPoolOrder(),
			window: window,
			state:  &memStore{},
		}
	})
	return defaultPool
}

// buildPoolOrder lists the available providers in priority order: Gemini first, OpenRouter
// second, skipping any whose API key is not configured.
func buildPoolOrder() []Classifier {
	var order []Classifier
	if configs.TrackerConfig.System.GeminiKey != "" {
		order = append(order, ClassifierGemini)
	}
	if configs.TrackerConfig.System.OpenRouterKey != "" {
		order = append(order, ClassifierOpenRouter)
	}
	return order
}

// sequence advances the sticky window and returns the providers to try for one request,
// ordered from the currently active provider so callers can fail over to the rest on error.
func (p *providerPool) sequence() []Classifier {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := len(p.order)
	if n == 0 {
		return nil
	}

	idx, used := p.state.load()
	if used >= p.window {
		idx = (idx + 1) % n
		used = 0
	}
	used++
	p.state.store(idx, used)

	seq := make([]Classifier, 0, n)
	for i := 0; i < n; i++ {
		seq = append(seq, p.order[(idx+i)%n])
	}
	return seq
}

// markRateLimited advances the cursor to the next provider so subsequent requests start there,
// after the active provider returned a rate-limit error.
func (p *providerPool) markRateLimited() {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := len(p.order)
	if n == 0 {
		return
	}
	idx, _ := p.state.load()
	p.state.store((idx+1)%n, 0)
}

// isRateLimited reports whether an error looks like a provider rate-limit / quota response.
func isRateLimited(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	for _, m := range []string{
		"rate limit", "ratelimit", "429", "quota",
		"resource_exhausted", "too many requests", "overloaded", "503",
	} {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}
