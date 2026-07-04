package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
)

// benchCase is one labeled input for the model benchmark.
type benchCase struct {
	Input  string
	Sub    string // expected subcategory ID
	Intent string // expected intent: income | expense | transfer
}

// benchFixtures: 50 labeled transaction inputs spanning the taxonomy, including the
// subject-aware debt cases. Used to score model accuracy and latency.
var benchFixtures = []benchCase{
	{"lunch at a restaurant 450", "food-rest", "expense"},
	{"grocery shopping for the week 1200", "food-groc", "expense"},
	{"bought fresh vegetables 300", "food-veg", "expense"},
	{"fish from the market 600", "food-fish", "expense"},
	{"chicken for dinner 550", "food-meat", "expense"},
	{"milk and eggs 180", "food-dairy", "expense"},
	{"cake from the bakery 400", "food-bakery", "expense"},
	{"fuchka and chotpoti 150", "food-street", "expense"},
	{"foodpanda delivery order 700", "food-take", "expense"},
	{"chips and biscuits 120", "food-snack", "expense"},
	{"coffee at the cafe 250", "food-bev", "expense"},
	{"bus fare to campus 40", "trans-pub", "expense"},
	{"uber ride to office 320", "trans-taxi", "expense"},
	{"petrol for the car 2000", "trans-fuel", "expense"},
	{"parking fee at mall 100", "trans-toll", "expense"},
	{"bike servicing and oil change 1500", "trans-maint", "expense"},
	{"new shirt for eid 1800", "shop-cloth", "expense"},
	{"bought a pair of sandals 900", "shop-foot", "expense"},
	{"phone charger and cable 600", "shop-elec", "expense"},
	{"lipstick and makeup 800", "shop-beauty", "expense"},
	{"notebook and pens 200", "shop-stat", "expense"},
	{"monthly salary credited 60000", "fin-sal", "income"},
	{"freelance project payment received 15000", "fin-prof", "income"},
	{"bank savings interest 500", "fin-interest", "income"},
	{"atm cash withdrawal 5000", "fin-with", "transfer"},
	{"deposited cash into bank account 10000", "fin-deposit", "transfer"},
	{"bkash send money to my other account 2000", "fin-transfer", "transfer"},
	{"mobile recharge flexiload 200", "fin-flexi", "expense"},
	{"credit card bill payment 12000", "fin-ccpay", "expense"},
	{"paid monthly house rent 15000", "house-rent", "expense"},
	{"electricity bill 1800", "house-util", "expense"},
	{"broadband internet bill 1000", "house-net", "expense"},
	{"maid monthly salary 3000", "house-serv", "expense"},
	{"doctor consultation fee 800", "health-doc", "expense"},
	{"blood test at the lab 1200", "health-test", "expense"},
	{"medicine from the pharmacy 450", "health-med", "expense"},
	{"haircut at the salon 300", "pc-salon", "expense"},
	{"gym membership fee 2000", "pc-fit", "expense"},
	{"netflix monthly subscription 500", "ent-sub", "expense"},
	{"movie ticket at the cinema 400", "ent-movie", "expense"},
	{"online programming course 3000", "edu-course", "expense"},
	{"flight ticket to cox bazar 8000", "trv-ticket", "expense"},
	{"hotel booking for the trip 6000", "trv-hotel", "expense"},
	{"eid shopping for the family 5000", "fest-eid", "expense"},
	{"zakat donation 5000", "fest-charity", "expense"},
	{"lent 3000 to rahim", "fin-lend", "expense"},
	{"borrowed 2000 from karim", "fin-borrow", "income"},
	{"john returned the 1500 he owed", "fin-recover", "income"},
	{"paid back 1000 to a friend", "fin-return", "expense"},
	{"friend paid me back 2500", "fin-recover", "income"},
}

// modelResult accumulates scoring for a single model.
type modelResult struct {
	provider    string
	model       string
	subCorrect  int
	intentMatch int
	failures    int
	latencies   []time.Duration
}

func (r *modelResult) attempted() int { return r.subCorrect + r.intentMatch }

// TestModelBenchmark scores every configured model on accuracy and latency. It hits live
// APIs, so it is gated behind RUN_AI_BENCH=1 and needs OPENROUTER_API_KEY / GEMINI_API_KEY.
//
//	set -a; source .env; set +a
//	RUN_AI_BENCH=1 go test -run TestModelBenchmark -v -timeout 60m ./modules/ai/
//
// Optional env: BENCH_LIMIT (cap fixtures), BENCH_MODELS (comma substrings to include),
// BENCH_DELAY_MS (pause between calls, default 800).
func TestModelBenchmark(t *testing.T) {
	if os.Getenv("RUN_AI_BENCH") == "" {
		t.Skip("set RUN_AI_BENCH=1 to run the live model benchmark")
	}

	taxonomyJSON := marshalTaxonomy(t)
	fixtures := limitedFixtures()
	delay := benchDelay()
	filter := os.Getenv("BENCH_MODELS")

	var results []*modelResult
	orKey := os.Getenv("OPENROUTER_API_KEY")
	if orKey != "" {
		client := NewClient(orKey)
		for _, m := range openRouterBenchModels() {
			if !modelIncluded(string(m), filter) {
				continue
			}
			results = append(results, runOpenRouterModel(t, client, m, taxonomyJSON, fixtures, delay))
		}
	} else {
		t.Log("OPENROUTER_API_KEY not set — skipping OpenRouter models")
	}

	gemKey := os.Getenv("GEMINI_API_KEY")
	if gemKey != "" {
		for _, m := range geminiBenchModels() {
			if !modelIncluded(m, filter) {
				continue
			}
			results = append(results, runGeminiModel(t, gemKey, m, taxonomyJSON, fixtures, delay))
		}
	} else {
		t.Log("GEMINI_API_KEY not set — skipping Gemini models")
	}

	if len(results) == 0 {
		t.Fatal("no models ran — check API keys and BENCH_MODELS filter")
	}
	reportBenchmark(t, results, len(fixtures))
}

func openRouterBenchModels() []OpenRouterModel {
	return []OpenRouterModel{
		DeepSeekChatFree, DeepSeekR1Free, DeepSeekR10525Free, Gemini20FlashFree,
		Gemma34bITFree, NVDIANemotron30bFree, NVDIANemotron9bFree, NVDIANemotron12bFree,
		DeepSeekV31NexN1Free, StepFun35FlashFree,
	}
}

func geminiBenchModels() []string {
	return []string{
		Gemini25FlashLite, Gemini20FlashLite, Gemini25Flash,
		Gemini3FlashPreview, Gemini35Flash, Gemini31FlashLite,
	}
}

func runOpenRouterModel(t *testing.T, client *Client, model OpenRouterModel, taxonomyJSON string, fixtures []benchCase, delay time.Duration) *modelResult {
	res := &modelResult{provider: "openrouter", model: string(model)}
	for _, fx := range fixtures {
		classify := func(ctx context.Context) (*ClassificationResult, error) {
			raw, err := client.query(ctx, model, client.buildPrompt(taxonomyJSON, fx.Input))
			if err != nil {
				return nil, err
			}
			var c ClassificationResult
			if err := json.Unmarshal([]byte(client.cleanJSON(raw)), &c); err != nil {
				return nil, err
			}
			return &c, nil
		}
		scoreOne(t, res, fx, classify)
		time.Sleep(delay)
	}
	return res
}

func runGeminiModel(t *testing.T, apiKey, model, taxonomyJSON string, fixtures []benchCase, delay time.Duration) *modelResult {
	res := &modelResult{provider: "gemini", model: model}
	for _, fx := range fixtures {
		classify := func(ctx context.Context) (*ClassificationResult, error) {
			return TxnSubcategoryClassifier(ctx, apiKey, fx.Input, taxonomyJSON, model)
		}
		scoreOne(t, res, fx, classify)
		time.Sleep(delay)
	}
	return res
}

// scoreOne runs a single classification with light retry on rate-limit/timeout and records
// the outcome (correctness + latency) into res.
func scoreOne(t *testing.T, res *modelResult, fx benchCase, classify func(context.Context) (*ClassificationResult, error)) {
	var out *ClassificationResult
	var err error
	var dur time.Duration
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		start := time.Now()
		out, err = classify(ctx)
		dur = time.Since(start)
		cancel()
		if err == nil {
			break
		}
		if !isRetryable(err) {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 3 * time.Second)
	}
	if err != nil {
		res.failures++
		t.Logf("[%s] %q FAILED: %v", res.model, fx.Input, trimErr(err))
		return
	}
	res.latencies = append(res.latencies, dur)
	subOK := out.Subcategory == fx.Sub
	intentOK := strings.EqualFold(out.Intent, fx.Intent)
	if subOK {
		res.subCorrect++
	}
	if intentOK {
		res.intentMatch++
	}
	if !subOK {
		t.Logf("[%s] %q -> %s/%s (want %s/%s)", res.model, fx.Input, out.Intent, out.Subcategory, fx.Intent, fx.Sub)
	}
}

func reportBenchmark(t *testing.T, results []*modelResult, total int) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].subCorrect > results[j].subCorrect
	})
	var b strings.Builder
	fmt.Fprintf(&b, "\n=== Model Benchmark (%d fixtures) ===\n", total)
	fmt.Fprintf(&b, "%-42s %-10s %-8s %-8s %-6s %8s %8s %8s\n",
		"MODEL", "PROVIDER", "SUB-ACC", "INT-ACC", "FAIL", "MEAN", "P50", "P95")
	for _, r := range results {
		graded := len(r.latencies)
		mean, p50, p95 := latencyStats(r.latencies)
		fmt.Fprintf(&b, "%-42s %-10s %-8s %-8s %-6d %8s %8s %8s\n",
			r.model, r.provider,
			pct(r.subCorrect, total), pct(r.intentMatch, total), r.failures,
			ms(mean), ms(p50), ms(p95))
		_ = graded
	}
	t.Log(b.String())
}

// --- helpers ---

func marshalTaxonomy(t *testing.T) string {
	data, err := json.MarshalIndent(models.TxnSubcategories, "", "  ")
	if err != nil {
		t.Fatalf("marshal taxonomy: %v", err)
	}
	return string(data)
}

func limitedFixtures() []benchCase {
	if v := os.Getenv("BENCH_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n < len(benchFixtures) {
			return benchFixtures[:n]
		}
	}
	return benchFixtures
}

func benchDelay() time.Duration {
	if v := os.Getenv("BENCH_DELAY_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Millisecond
		}
	}
	return 800 * time.Millisecond
}

func modelIncluded(model, filter string) bool {
	if filter == "" {
		return true
	}
	for _, want := range strings.Split(filter, ",") {
		if want = strings.TrimSpace(want); want != "" && strings.Contains(model, want) {
			return true
		}
	}
	return false
}

func isRetryable(err error) bool {
	msg := strings.ToLower(err.Error())
	for _, s := range []string{"429", "rate", "timeout", "deadline", "too many", "503", "overload"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func trimErr(err error) string {
	s := err.Error()
	if len(s) > 160 {
		return s[:160] + "…"
	}
	return s
}

func latencyStats(ds []time.Duration) (mean, p50, p95 time.Duration) {
	if len(ds) == 0 {
		return 0, 0, 0
	}
	sorted := append([]time.Duration(nil), ds...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	var sum time.Duration
	for _, d := range sorted {
		sum += d
	}
	mean = sum / time.Duration(len(sorted))
	p50 = sorted[len(sorted)*50/100]
	p95 = sorted[min(len(sorted)*95/100, len(sorted)-1)]
	return mean, p50, p95
}

func pct(n, total int) string {
	if total == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.0f%%", 100*float64(n)/float64(total))
}

func ms(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}
