package transaction

import (
	"fmt"
	"strings"

	"github.com/masudur-rahman/expense-tracker-bot/models"
)

// Debt transactions always involve a person and a direction of money. The four debt
// subcategories are two axes: money direction (in/out) × new-loan vs. settlement.
//
//	fin-lend    out + new loan        (I give a loan)
//	fin-recover in  + settlement      (they return what I lent)
//	fin-borrow  in  + new loan        (I take a loan)
//	fin-return  out + settlement      (I give back what I borrowed)
//
// The verb tells direction, "back"-family words mark settlement, and the subject
// (who acted on whom) can flip the direction — "John returned" is money coming to me.

// Multi-word and Banglish debt phrases are collapsed into single tokens so the word-level
// logic below can match them (the tokenizer only sees one word at a time). Matching is done
// per whole word, so "back from" does not fire inside "cashback from".
var (
	debtTrigrams = map[string]string{"paid me back": "paidmeback"}
	debtBigrams  = map[string]string{
		"paid back": "paidback", "pay back": "paidback",
		"gave back": "gaveback", "give back": "gaveback",
		"got back": "gotback", "get back": "getback", "back from": "getback",
		"took loan": "borrowed", "gave loan": "lent", "paid for": "paidfor",
		"dhar dilam": "lent", "dhar disi": "lent",
		"dhar nilam": "borrowed", "dhar nisi": "borrowed",
		"dhar shodh": "repaid", "dhar shod": "repaid",
		"ferot pelam": "getback", "ferot dilam": "returned", "ferot dilo": "returned",
	}
	debtUnigrams = map[string]string{"payback": "paidback"}
)

// canonicalizeDebt tokenizes text and merges known debt phrases into single tokens.
func canonicalizeDebt(text string) []string {
	words := strings.Fields(text)
	out := make([]string, 0, len(words))
	for i := 0; i < len(words); i++ {
		if i+2 < len(words) {
			if v, ok := debtTrigrams[words[i]+" "+words[i+1]+" "+words[i+2]]; ok {
				out, i = append(out, v), i+2
				continue
			}
		}
		if i+1 < len(words) {
			if v, ok := debtBigrams[words[i]+" "+words[i+1]]; ok {
				out, i = append(out, v), i+1
				continue
			}
		}
		if v, ok := debtUnigrams[words[i]]; ok {
			out = append(out, v)
			continue
		}
		out = append(out, words[i])
	}
	return out
}

var (
	debtOutVerbs = map[string]bool{
		"gave": true, "give": true, "given": true, "lent": true, "lend": true,
		"handed": true, "paid": true, "pay": true, "paidfor": true, "covering": true,
		"cover": true, "disi": true, "dilam": true, "returned": true, "repaid": true,
		"gaveback": true, "paidback": true, "paidmeback": true,
	}
	debtInVerbs = map[string]bool{
		"got": true, "get": true, "received": true, "receive": true, "recovered": true,
		"recover": true, "collected": true, "collect": true, "took": true, "take": true,
		"borrowed": true, "borrow": true, "pelam": true, "paisi": true, "nilam": true,
		"nisi": true, "gotback": true, "getback": true, "cashback": true,
	}
	settlementTokens = map[string]bool{
		"gotback": true, "getback": true, "returned": true, "repaid": true,
		"gaveback": true, "paidback": true, "paidmeback": true, "ferot": true,
		"shod": true, "shodh": true, "back": true, "repayment": true, "lending": true,
	}
	// strongDebtTokens unambiguously mark a debt transaction on their own. Generic verbs
	// (paid/got/gave/took) are NOT here — they're debt only in a bare person+amount context,
	// so "paid wifi bill", "got bonus", "paid masud for lunch" are left to normal classification.
	strongDebtTokens = map[string]bool{
		"lent": true, "lend": true, "borrowed": true, "borrow": true, "loan": true,
		"dhar": true, "recovered": true, "recover": true, "collected": true,
		"collect": true, "repaid": true, "returned": true, "ferot": true, "shod": true,
		"shodh": true, "covering": true, "gaveback": true, "paidback": true,
		"paidmeback": true, "gotback": true, "getback": true, "nilam": true, "nisi": true,
		"pelam": true, "disi": true, "dilam": true, "repayment": true, "lending": true,
		"cashback": true,
	}
	// personWords are generic references to a person, letting the resolver engage and
	// record a [Person: ...] marker even when no contact is on file.
	personWords = map[string]bool{
		"friend": true, "friends": true, "colleague": true, "guard": true, "maid": true,
		"driver": true, "shopkeeper": true, "batchmate": true, "someone": true,
		"bhai": true, "vai": true, "apu": true, "mama": true, "chacha": true,
	}
	// subjectStopwords are never treated as the acting person (subject) or a [Person:].
	subjectStopwords = map[string]bool{
		"i": true, "we": true, "me": true, "my": true, "him": true, "her": true,
		"them": true, "back": true, "the": true, "a": true, "to": true, "from": true,
		"ke": true, "theke": true, "re": true, "for": true, "of": true,
	}
	// objectMarkers follow a person to mark them as the recipient (object), not subject —
	// crucial for Banglish SOV order ("John ke ... disi" = gave TO John).
	objectMarkers = map[string]bool{"ke": true, "to": true, "re": true}
)

// classifyDebt resolves a debt subcategory from the (canonicalized, lower-cased) text, or
// returns false when the input is not debt-related.
func classifyDebt(text string, isContact ContactVerifier) (string, bool) {
	words := canonicalizeDebt(text)
	verbIdx, dir, matched := primaryDebtVerb(words)

	// Engage only on a strong debt token, or a bare "person + amount" reference. A generic
	// verb next to a real category ("paid masud for lunch") is a normal transaction.
	if !hasStrongDebt(words) && !bareContactContext(words, isContact) {
		return "", false
	}
	if !matched {
		// Person/signal present but no direction verb ("John 500") — default to lending.
		return models.LendSubID, true
	}

	settlement := hasSettlement(words)
	if personActedOnMe(words, verbIdx, isContact) {
		dir = moneyIn
	}
	return debtSubcategory(dir, settlement), true
}

// hasStrongDebt reports whether any token unambiguously marks a debt transaction.
func hasStrongDebt(words []string) bool {
	for _, w := range words {
		if strongDebtTokens[w] {
			return true
		}
	}
	return false
}

type moneyDir int

const (
	moneyOut moneyDir = iota
	moneyIn
)

func debtSubcategory(dir moneyDir, settlement bool) string {
	switch {
	case dir == moneyIn && settlement:
		return models.LendRecoverySubID
	case dir == moneyIn:
		return models.BorrowSubID
	case settlement:
		return models.BorrowReturnSubID
	default:
		return models.LendSubID
	}
}

// debtFillers are connector/quantity words that carry no category meaning, so a sentence
// made only of them plus a person and an amount is a bare person reference.
var debtFillers = map[string]bool{
	"with": true, "transaction": true, "money": true, "cash": true, "taka": true,
	"tk": true, "bdt": true, "and": true, "the": true, "a": true, "an": true,
	"to": true, "from": true, "for": true, "of": true, "ke": true, "theke": true,
	"re": true, "some": true, "just": true, "give": true, "given": true,
}

// bareContactContext reports whether, after removing the person and amount, only filler
// words remain — i.e. the input is just a person + amount with no other category signal.
func bareContactContext(words []string, isContact ContactVerifier) bool {
	seenPerson := false
	for _, w := range words {
		switch {
		case isContact(w) || personWords[w]:
			seenPerson = true
		case isNumericToken(w) || debtFillers[w] || debtInVerbs[w] || debtOutVerbs[w]:
			// ignore amounts, connectors, and the debt verb itself
		default:
			return false
		}
	}
	return seenPerson
}

func isNumericToken(w string) bool {
	if w == "" {
		return false
	}
	for _, r := range w {
		if (r < '0' || r > '9') && r != '.' && r != 'k' {
			return false
		}
	}
	return true
}

// primaryDebtVerb returns the index and money direction of the first debt verb found.
func primaryDebtVerb(words []string) (int, moneyDir, bool) {
	for i, w := range words {
		if debtInVerbs[w] {
			return i, moneyIn, true
		}
		if debtOutVerbs[w] {
			return i, moneyOut, true
		}
	}
	return 0, moneyOut, false
}

func hasSettlement(words []string) bool {
	for _, w := range words {
		if settlementTokens[w] {
			return true
		}
	}
	return false
}

// personActedOnMe reports whether another person is the subject of the verb (so money
// flows to me): a person token before the verb that is not marked as a recipient, or an
// explicit "...me back"/"paid me" object.
func personActedOnMe(words []string, verbIdx int, isContact ContactVerifier) bool {
	for i := 0; i < verbIdx; i++ {
		w := words[i]
		if subjectStopwords[w] {
			continue
		}
		if !isContact(w) && !personWords[w] {
			continue
		}
		if i+1 < len(words) && objectMarkers[words[i+1]] {
			continue // "John ke ..." — John is the recipient, not the actor
		}
		return true
	}
	return words[verbIdx] == "paidmeback"
}

// resolveDebtDirection sets the debt subcategory/type and attaches the person when the
// input is debt-related. A no-op otherwise, leaving normal classification untouched.
func (p *transactionParser) resolveDebtDirection(isContact ContactVerifier) {
	subID, ok := classifyDebt(p.rawText, isContact)
	if !ok {
		return
	}
	p.subcategory = subID
	if subID == models.LendSubID || subID == models.BorrowReturnSubID {
		p.txnType = models.ExpenseTransaction
	} else {
		p.txnType = models.IncomeTransaction
	}
	p.attachDebtPerson(isContact)
}

// attachDebtPerson records the person involved: a known contact becomes ContactName
// (driving balance updates); an unknown person is left as a [Person: name] remark so a
// later crawler can reconcile it. Persons already captured by from/to are left to
// finalizeMapping.
func (p *transactionParser) attachDebtPerson(isContact ContactVerifier) {
	person := p.findDebtPerson(isContact)
	if person == "" || person == p.toValue || person == p.fromValue {
		return
	}
	if isContact(person) {
		if p.txn.ContactName == "" {
			p.txn.ContactName = person
		}
		return
	}
	marker := fmt.Sprintf("[Person: %s]", person)
	if !strings.Contains(p.note, marker) {
		p.appendNote(marker)
	}
}

// findDebtPerson picks the person referenced in a debt sentence: a known contact first,
// else a generic person word, skipping stopwords/pronouns.
func (p *transactionParser) findDebtPerson(isContact ContactVerifier) string {
	words := canonicalizeDebt(p.rawText)
	for _, w := range words {
		if isContact(w) {
			return w
		}
	}
	for _, w := range words {
		if personWords[w] && !subjectStopwords[w] {
			return w
		}
	}
	return ""
}
