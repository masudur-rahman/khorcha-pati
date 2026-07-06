package transaction

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/ai"

	"go.uber.org/zap"
)

// benchCase is one labeled input for the end-to-end flow benchmark. Inputs are curated to
// bypass the verb-keyword, debt and local-keyword layers so they reach the AI classifier;
// the pre-pass in TestFlowBenchmark flags any that don't.
type benchCase struct {
	Input string
	Sub   string                 // expected final subcategory ID
	Type  models.TransactionType // expected transaction type
}

// benchFixtures are realistic everyday entries a Bangladeshi user would actually type — short,
// casual, a mix of English and Banglish. This exercises the whole pipeline: common Banglish now
// resolves in localClassify (caught locally, no AI call) while long-tail and indirect phrasing
// reaches the AI. The pre-pass reports the split; expected labels are the correct final result
// either way.
var benchFixtures = []benchCase{
	// --- English, casual (how a Bangladeshi jots it) ---
	{"printout and photocopy 150", "shop-stat", models.ExpenseTransaction},
	{"physiotherapy session 2200", "health-other", models.ExpenseTransaction},
	{"birthday bouquet for mom 600", "misc-gift", models.ExpenseTransaction},
	{"annual eye checkup 700", "health-other", models.ExpenseTransaction},
	{"farewell present for a colleague 1200", "misc-gift", models.ExpenseTransaction},
	{"treat for my friends 600", "food-rest", models.ExpenseTransaction},
	{"monthly consultancy retainer 25000", "fin-prof", models.IncomeTransaction},
	{"rahim owes me 2000", "fin-lend", models.ExpenseTransaction},
	{"I owe karim 3000", "fin-borrow", models.IncomeTransaction},
	// --- Banglish ---
	{"gari mekanik 1200", "trans-maint", models.ExpenseTransaction},
	{"chuler kat 250", "pc-salon", models.ExpenseTransaction},
	{"daater daktar 800", "health-other", models.ExpenseTransaction},
	{"taja maach 600", "food-fish", models.ExpenseTransaction},
	{"leguna vara 30", "trans-pub", models.ExpenseTransaction},
	{"bhai er jonno jama 900", "shop-cloth", models.ExpenseTransaction},
	{"notun juta 1200", "shop-foot", models.ExpenseTransaction},
	{"basar bajar 2000", "food-groc", models.ExpenseTransaction},
	{"cha nasta 120", "food-snack", models.ExpenseTransaction},
	{"bank theke rin 50000", "fin-loan", models.IncomeTransaction},
	{"karim er dena clear korlam 1500", "fin-return", models.ExpenseTransaction},
	{"rahim er theke pawna tullam 2500", "fin-recover", models.IncomeTransaction},
}

// benchModel identifies one provider+model to run through the flow.
type benchModel struct {
	provider string // "gemini" | "open-router"
	model    string
}

// benchModelList is the set of models compared. Filter at runtime with BENCH_MODELS.
func benchModelList() []benchModel {
	var ms []benchModel
	for _, m := range []ai.OpenRouterModel{
		ai.NVDIANemotron30bFree,
		//ai.NVIDIANemotronSuper120bFree,
		//ai.OpenAIGPTOSS120bFree,
	} {
		ms = append(ms, benchModel{"open-router", string(m)})
	}
	for _, m := range []string{
		ai.Gemini31FlashLite,
		//ai.Gemini35Flash,
	} {
		ms = append(ms, benchModel{"gemini", m})
	}
	return ms
}

// modelResult accumulates scoring for a single model.
type modelResult struct {
	provider   string
	model      string
	subCorrect int
	typeMatch  int
	failures   int
	latencies  []time.Duration
}

// TestFlowBenchmark runs each fixture through the full ParseTransaction pipeline for every
// configured model and scores the final subcategory + type. It hits live APIs, so it is gated
// behind RUN_AI_BENCH=1 and needs OPENROUTER_API_KEY / GEMINI_API_KEY.
//
//	set -a; source .env; set +a
//	RUN_AI_BENCH=1 go test -run TestFlowBenchmark -v -timeout 60m ./modules/transaction/
//
// Optional env: BENCH_LIMIT (cap fixtures), BENCH_MODELS (comma substrings to include),
// BENCH_DELAY_MS (pause between calls, default 800).
func TestFlowBenchmark(t *testing.T) {
	if os.Getenv("RUN_AI_BENCH") == "" {
		t.Skip("set RUN_AI_BENCH=1 to run the live end-to-end benchmark")
	}

	// The flow mutates package-global config and logs DB-persist errors (no DB here). Save and
	// restore config, and silence the logger for the run so the report stays readable.
	origSys := configs.TrackerConfig.System
	origLog := logr.DefaultLogger
	defer func() {
		configs.TrackerConfig.System = origSys
		logr.DefaultLogger = origLog
	}()
	logr.DefaultLogger = zap.NewNop().Sugar()

	fixtures := limitedFixtures()

	// Pre-pass: confirm every fixture actually reaches the AI classifier.
	reportAIReach(t, fixtures)

	gemKey := os.Getenv("GEMINI_API_KEY")
	orKey := os.Getenv("OPENROUTER_API_KEY")
	filter := os.Getenv("BENCH_MODELS")
	delay := benchDelay()

	var results []*modelResult
	for _, m := range benchModelList() {
		if !modelIncluded(m.model, filter) {
			continue
		}
		if m.provider == "gemini" && gemKey == "" {
			continue
		}
		if m.provider == "open-router" && orKey == "" {
			continue
		}
		applyModelConfig(m, gemKey, orKey)
		results = append(results, runModel(t, m, fixtures, delay))
	}

	if len(results) == 0 {
		t.Fatal("no models ran — check API keys and BENCH_MODELS filter")
	}
	reportBenchmark(t, results, len(fixtures))
}

// reportAIReach runs each fixture with the AI disabled: if the local layers (verb keywords,
// debt resolver, cache, localClassify) resolve it to a formal subcategory, it will never reach
// the AI and is flagged; otherwise the leftover phrase survives, confirming it is AI-bound.
func reportAIReach(t *testing.T, fixtures []benchCase) {
	sys := &configs.TrackerConfig.System
	sys.AIClassifier, sys.GeminiKey, sys.OpenRouterKey = "", "", ""

	formal := formalSubcategoryIDs()
	var caught int
	var b strings.Builder
	b.WriteString("\n=== AI-reach pre-pass (AI disabled) ===\n")
	for _, fx := range fixtures {
		txn, err := ParseTransaction(fx.Input, noVerifier, noVerifier)
		if err != nil {
			fmt.Fprintf(&b, "ERROR %-58q parse failed: %v\n", fx.Input, err)
			continue
		}
		if formal[txn.SubcategoryID] {
			caught++
			fmt.Fprintf(&b, "LOCAL %-58q -> %s (won't hit AI; expected %s)\n", fx.Input, txn.SubcategoryID, fx.Sub)
		} else {
			fmt.Fprintf(&b, "AI    %-58q -> leftover %q\n", fx.Input, txn.SubcategoryID)
		}
	}
	fmt.Fprintf(&b, "%d/%d reach AI; %d caught locally\n", len(fixtures)-caught, len(fixtures), caught)
	t.Log(b.String())
}

func runModel(t *testing.T, m benchModel, fixtures []benchCase, delay time.Duration) *modelResult {
	res := &modelResult{provider: m.provider, model: m.model}
	for _, fx := range fixtures {
		start := time.Now()
		txn, err := ParseTransaction(fx.Input, noVerifier, noVerifier)
		dur := time.Since(start)
		if err != nil {
			res.failures++
			t.Logf("[%s] %q FAILED: %v", m.model, fx.Input, trimErr(err))
			time.Sleep(delay)
			continue
		}
		res.latencies = append(res.latencies, dur)
		if txn.SubcategoryID == fx.Sub {
			res.subCorrect++
		}
		if txn.Type == fx.Type {
			res.typeMatch++
		}
		if txn.SubcategoryID != fx.Sub {
			t.Logf("[%s] %q -> %s/%s (want %s/%s)", m.model, fx.Input, txn.Type, txn.SubcategoryID, fx.Type, fx.Sub)
		}
		time.Sleep(delay)
	}
	return res
}

// applyModelConfig points the flow at a single provider+model. The pool is bypassed so exactly
// one model runs per iteration.
func applyModelConfig(m benchModel, gemKey, orKey string) {
	sys := &configs.TrackerConfig.System
	sys.AIClassifier = m.provider
	if m.provider == "gemini" {
		sys.GeminiKey = gemKey
		sys.GeminiModel = m.model
	} else {
		sys.OpenRouterKey = orKey
		sys.OpenRouterModel = m.model
	}
}

func reportBenchmark(t *testing.T, results []*modelResult, total int) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].subCorrect > results[j].subCorrect
	})
	var b strings.Builder
	fmt.Fprintf(&b, "\n=== Flow Benchmark (%d fixtures, full pipeline) ===\n", total)
	fmt.Fprintf(&b, "%-42s %-12s %-8s %-9s %-6s %8s %8s %8s\n",
		"MODEL", "PROVIDER", "SUB-ACC", "TYPE-ACC", "FAIL", "MEAN", "P50", "P95")
	for _, r := range results {
		mean, p50, p95 := latencyStats(r.latencies)
		fmt.Fprintf(&b, "%-42s %-12s %-8s %-9s %-6d %8s %8s %8s\n",
			r.model, r.provider,
			pct(r.subCorrect, total), pct(r.typeMatch, total), r.failures,
			ms(mean), ms(p50), ms(p95))
	}
	t.Log(b.String())
}

// --- helpers ---

func noVerifier(string) bool { return false }

func formalSubcategoryIDs() map[string]bool {
	ids := make(map[string]bool, len(models.TxnSubcategories))
	for _, sub := range models.TxnSubcategories {
		ids[sub.ID] = true
	}
	return ids
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
