package transaction

import (
	"strings"

	"github.com/masudur-rahman/expense-tracker-bot/models"
)

// fillerWords are dropped when normalizing a phrase for cache lookup so that
// "had lunch", "a lunch", and "lunch" collapse to the same cache key.
var fillerWords = map[string]bool{
	"a": true, "an": true, "the": true, "some": true,
	"had": true, "have": true, "my": true, "of": true,
}

// normalizePhrase lower-cases, strips punctuation and filler words, and collapses
// whitespace so equivalent inputs share one cache key (raising the AI cache hit rate).
func normalizePhrase(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == ' ':
			b.WriteRune(r)
		case r >= 0x0980 && r <= 0x09FF: // Bengali block — keep for Banglish inputs
			b.WriteRune(r)
		default:
			b.WriteRune(' ')
		}
	}
	fields := make([]string, 0)
	for _, w := range strings.Fields(b.String()) {
		if fillerWords[w] {
			continue
		}
		fields = append(fields, w)
	}
	return strings.Join(fields, " ")
}

// localClassify matches a phrase against subcategory keywords/names locally, without
// calling the AI. It returns the best subcategory ID when a whole-word keyword hits,
// preferring more specific (longer, multi-word) keywords. This keeps the common case
// off the rate-limited AI endpoint.
func localClassify(phrase string) (string, bool) {
	phrase = normalizePhrase(phrase)
	if phrase == "" {
		return "", false
	}
	padded := " " + phrase + " "

	bestID := ""
	bestScore := 0
	for _, sub := range models.TxnSubcategories {
		for _, kw := range subcategoryKeywords(sub) {
			if !strings.Contains(padded, " "+kw+" ") {
				continue
			}
			// Longer, multi-word keywords are more specific → higher score.
			score := len(strings.Fields(kw))*100 + len(kw)
			if score > bestScore {
				bestScore, bestID = score, sub.ID
			}
		}
	}
	if bestID == "" {
		return "", false
	}
	return bestID, true
}

// subcategoryKeywords returns the normalized whole-word search terms for a subcategory:
// its comma-separated Keywords plus its Name.
func subcategoryKeywords(sub models.TxnSubcategory) []string {
	terms := make([]string, 0)
	for _, kw := range strings.Split(sub.Keywords, ",") {
		if kw = normalizePhrase(kw); kw != "" {
			terms = append(terms, kw)
		}
	}
	if name := normalizePhrase(sub.Name); name != "" {
		terms = append(terms, name)
	}
	return terms
}

// isRateLimitErr reports whether an AI error is a quota/rate-limit failure, so the
// caller can degrade gracefully instead of dropping the user's transaction.
func isRateLimitErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{"rate limit", "ratelimit", "429", "quota", "resource_exhausted", "too many requests"} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}
