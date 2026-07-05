package ai

import (
	"errors"
	"testing"
)

// newTestPool builds a pool with a fixed order and window, bypassing config.
func newTestPool(window int, order ...GeneratorAI) *providerPool {
	return &providerPool{order: order, window: window, state: &memStore{}}
}

// TestProviderPool_sticky asserts the pool serves one provider for a full window, then rotates.
func TestProviderPool_sticky(t *testing.T) {
	p := newTestPool(3, GeneratorGemini, GeneratorOpenRouter)

	for i := 0; i < 3; i++ {
		if got := p.sequence()[0]; got != GeneratorGemini {
			t.Fatalf("request %d: primary = %q, want gemini", i, got)
		}
	}
	// window exhausted → rotate to OpenRouter
	if got := p.sequence()[0]; got != GeneratorOpenRouter {
		t.Fatalf("after window: primary = %q, want open-router", got)
	}
}

// TestProviderPool_sequenceOrder asserts sequence lists all providers for failover, active first.
func TestProviderPool_sequenceOrder(t *testing.T) {
	p := newTestPool(10, GeneratorGemini, GeneratorOpenRouter)
	seq := p.sequence()
	if len(seq) != 2 || seq[0] != GeneratorGemini || seq[1] != GeneratorOpenRouter {
		t.Fatalf("sequence = %v, want [gemini open-router]", seq)
	}
}

// TestProviderPool_markRateLimited asserts a rate-limit advances the active provider.
func TestProviderPool_markRateLimited(t *testing.T) {
	p := newTestPool(10, GeneratorGemini, GeneratorOpenRouter)
	_ = p.sequence() // active = gemini
	p.markRateLimited()
	if got := p.sequence()[0]; got != GeneratorOpenRouter {
		t.Fatalf("after rate-limit: primary = %q, want open-router", got)
	}
}

// TestProviderPool_singleProvider asserts a one-provider pool never rotates and has no failover.
func TestProviderPool_singleProvider(t *testing.T) {
	p := newTestPool(2, GeneratorGemini)
	for i := 0; i < 5; i++ {
		seq := p.sequence()
		if len(seq) != 1 || seq[0] != GeneratorGemini {
			t.Fatalf("request %d: seq = %v, want [gemini]", i, seq)
		}
	}
}

// TestProviderPool_empty asserts an unconfigured pool yields no providers.
func TestProviderPool_empty(t *testing.T) {
	p := newTestPool(3)
	if seq := p.sequence(); seq != nil {
		t.Fatalf("sequence = %v, want nil", seq)
	}
	p.markRateLimited() // must not panic
}

// TestIsRateLimited checks the rate-limit error matcher.
func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("API error 429: too many requests"), true},
		{errors.New("RESOURCE_EXHAUSTED: quota"), true},
		{errors.New("model overloaded"), true},
		{errors.New("API error 503"), true},
		{errors.New("failed to parse JSON"), false},
		{errors.New("connection refused"), false},
	}
	for _, tt := range tests {
		if got := isRateLimited(tt.err); got != tt.want {
			t.Errorf("isRateLimited(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}
