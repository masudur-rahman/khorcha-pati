package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/ai"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
	"github.com/masudur-rahman/khorcha-pati/pkg"
)

/*
	transaction examples: [any keyword can be anywhere, like in natural language]
		transfer 2000 from brac to dbbl on 2020-01-01 note "Bill payment"
		spend 1000 for food-rest on "Jan 13, 2013" from dbbl note "Lunch"
		earn 5000 to brac on 2020-01-01 note "Salary"
		borrow 1000 from user to brac on 2020-01-01
		return 1000 to user from brac on 2020-01-01
		lend 1000 to user from brac on 2020-01-01
		recover 1000 from user to brac on 2020-01-01

	verb keywords: (<keyword> <amount>)
		- transfer
		- expense, spend
		- income, earn
		- borrow
		- return
		- lend
		- recover

	other keywords:
		- from 	[source wallet (for normal transaction)] [person (for borrow and recover)]
		- to 	[destination wallet (for normal transaction)] [person (for lend and return)]
		- for 	[subcategory]
		- on 	[date]
		- at	[time]
		- note	[note]
*/

// Verifiers for dependency injection
type ContactVerifier func(name string) bool

type AccountVerifier func(name string) bool

type transactionParser struct {
	txn         models.Transaction
	txnType     models.TransactionType
	amount      string
	fromValue   string
	toValue     string
	subcategory string
	date        string
	time        string
	note        string
	rawText     string
	verbFound   bool
	// bareAccounts / bareContacts are wallet and contact names mentioned
	// without a preposition ("salary 50k ebl", "lent 500 karim") — routed
	// once the transaction type is known.
	bareAccounts []string
	bareContacts []string
	tz           *time.Location
}

func ParseTransaction(text string, isContact ContactVerifier, isAccount AccountVerifier, tz *time.Location) (models.Transaction, error) {
	if tz == nil {
		tz = pkg.DefaultLocation
	}
	// --- STEP 0: Safety Defaults ---
	if isContact == nil {
		isContact = func(name string) bool { return false }
	}
	if isAccount == nil {
		isAccount = func(name string) bool { return false }
	}

	p := transactionParser{tz: tz}

	// --- STEP 1: Find Amount ---
	// The thousands multiplier needs a word boundary so "500 karim" isn't read
	// as "500k" with a leftover "arim".
	reAmount := regexp.MustCompile(`(?i)(?:(?:total|tk|taka|amount)\s*)?(\d+(?:\.\d+)?)\s*(k\b)?(?:\s*(?:tk|taka|bdt))?`)
	reUnit := regexp.MustCompile(`(?i)^(?:kg|km|g|gm|lb|lbs|ml|mg|oz|l|pcs?|pieces?)\b`)
	matches := reAmount.FindAllStringSubmatchIndex(text, -1)

	loc := findMonetaryAmount(text, matches, reUnit)
	if loc == nil {
		return models.Transaction{}, fmt.Errorf("no valid amount found in text")
	}

	numberStr := text[loc[2]:loc[3]]
	if loc[4] != -1 {
		val, _ := strconv.ParseFloat(numberStr, 64)
		p.amount = fmt.Sprintf("%.2f", val*1000)
	} else {
		p.amount = numberStr
	}
	textWithoutAmount := text[:loc[0]] + " " + text[loc[1]:]
	p.rawText = strings.ToLower(strings.TrimSpace(textWithoutAmount))

	// --- STEP 2: Tokenize & Scan ---
	words := strings.Fields(textWithoutAmount)
	p.txnType = models.ExpenseTransaction

	var currentKey string
	var currentBuffer []string

	for i := 0; i < len(words); i++ {
		word := words[i]
		lowerWord := strings.ToLower(word)

		if lowerWord == "note" {
			p.flushBuffer(currentKey, currentBuffer, isContact, isAccount)
			if i+1 < len(words) {
				p.note = strings.Join(words[i+1:], " ")
			}
			currentKey = ""
			currentBuffer = nil
			break
		}
		if p.isVerbKeyword(lowerWord) {
			p.verbFound = true
			p.flushBuffer(currentKey, currentBuffer, isContact, isAccount)
			currentBuffer = []string{}
			currentKey = ""
			continue
		}
		if isDateKeyword(lowerWord) {
			p.flushBuffer(currentKey, currentBuffer, isContact, isAccount)
			p.date = lowerWord
			currentBuffer = []string{}
			currentKey = ""
			continue
		}
		// Named times of day ("dinner 400 night") work without an "at".
		if pkg.IsNamedHour(lowerWord) {
			p.flushBuffer(currentKey, currentBuffer, isContact, isAccount)
			p.time = lowerWord
			currentBuffer = []string{}
			currentKey = ""
			continue
		}
		if isStandardKeyword(lowerWord) {
			p.flushBuffer(currentKey, currentBuffer, isContact, isAccount)
			currentKey = lowerWord
			currentBuffer = []string{}
		} else {
			currentBuffer = append(currentBuffer, word)
		}
	}
	p.flushBuffer(currentKey, currentBuffer, isContact, isAccount)
	p.cleanSubcategory()

	// --- STEP 3: Resolve debt direction (subject-aware, pre-AI) ---
	p.resolveDebtDirection(isContact)

	// --- STEP 3.5: Enrich Context (Pre-AI) ---
	p.enrichContext(isAccount)

	// --- STEP 4: Resolve ID (AI/Cache) ---
	if err := p.subcategoryAIParser(); err != nil {
		return models.Transaction{}, err
	}

	// --- STEP 5: Finalize Mapping (Post-AI) ---
	p.finalizeMapping(isContact, isAccount)

	// --- STEP 6: Finalize Struct ---
	err := p.parseTransaction()
	return p.txn, err
}

// reTimeOrdinal matches time-of-day and ordinal fragments right after a
// number ("5pm", "5:30", "1st") — those are never monetary amounts.
var reTimeOrdinal = regexp.MustCompile(`(?i)^(?:\s*(?:am|pm)\b|(?:st|nd|rd|th)\b|:\d)`)

// findMonetaryAmount picks the best monetary amount from regex matches,
// skipping numbers followed by measurement units (kg, g, ml, etc.) and
// time/ordinal fragments ("5pm", "5:30", "1st").
func findMonetaryAmount(text string, matches [][]int, reUnit *regexp.Regexp) []int {
	// Prefer matches with currency suffix (tk/taka/bdt) or k multiplier
	var fallback []int
	for _, loc := range matches {
		after := text[loc[1]:]
		if reUnit.MatchString(after) || reTimeOrdinal.MatchString(after) {
			continue
		}
		// A number preceded by ":" is the minutes half of a clock time.
		if loc[2] > 0 && text[loc[2]-1] == ':' {
			continue
		}
		hasCurrency := loc[1] > loc[3] || (loc[4] != -1)
		if hasCurrency {
			return loc
		}
		if fallback == nil {
			fallback = loc
		}
	}
	return fallback
}

func (p *transactionParser) enrichContext(isAccount AccountVerifier) {
	if p.isFormalSubcategoryID() {
		return
	}

	mergeIntoText := func(val string, isSource bool) {
		if val == "" {
			return
		}

		if isAccount(strings.ToLower(val)) {
			return
		}

		prefix := "to"
		if isSource {
			prefix = "from"
		}
		info := fmt.Sprintf("%s %s", prefix, val)

		if p.subcategory != "" {
			p.subcategory += " " + info
		} else {
			p.subcategory = info
		}
	}

	mergeIntoText(p.fromValue, true)
	mergeIntoText(p.toValue, false)
}

func (p *transactionParser) subcategoryAIParser() error {
	for _, subcat := range models.TxnSubcategories {
		if subcat.ID == p.subcategory {
			return nil
		}
	}

	if p.note == "" {
		p.note = p.subcategory
	} else {
		p.note = p.subcategory + " " + p.note
	}

	p.subcategory = normalizePhrase(p.subcategory)
	if cached, exist := cache.GetCache(p.subcategory); exist {
		var result ai.ClassificationResult
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			p.subcategory = result.Subcategory
			p.setIntent(result.Intent)
			return nil
		}
		// Fallback for old cache format
		p.subcategory = cached
		return nil
	}

	// Keyword-first: resolve common inputs locally so the rate-limited AI endpoint
	// is only hit for genuinely ambiguous text.
	if subID, ok := localClassify(p.subcategory); ok {
		p.subcategory = subID
		return nil
	}

	inputText := p.subcategory
	result, err := ai.TxnCategoryClassifier(context.Background(), inputText)
	if err != nil {
		// Degrade gracefully on quota/rate-limit: never drop the user's transaction.
		if isRateLimitErr(err) {
			logr.DefaultLogger.Warnw("AI rate-limited, falling back to misc", "input", inputText)
			p.subcategory = "misc-misc"
			return nil
		}
		return err
	}

	// Only cache when AI actually classified (not a passthrough from missing API key).
	if result.Subcategory != "" && result.Subcategory != inputText {
		resultJSON, _ := json.Marshal(result)
		_ = cache.SetCache(inputText, string(resultJSON), -1)
		if dbErr := configs.InsertAICache(models.AICache{
			InputText:     inputText,
			SubcategoryID: result.Subcategory,
			Intent:        result.Intent,
			Confidence:    result.Confidence,
			CreatedAt:     time.Now().In(p.tz).Unix(),
		}); dbErr != nil {
			logr.DefaultLogger.Errorw("Failed to persist AI cache", "error", dbErr.Error())
		}
	}

	p.subcategory = result.Subcategory
	p.setIntent(result.Intent)
	return nil
}

func (p *transactionParser) setIntent(intent string) {
	if intent == "" {
		return
	}
	switch strings.ToLower(intent) {
	case "income":
		p.txnType = models.IncomeTransaction
	case "expense":
		p.txnType = models.ExpenseTransaction
	case "transfer":
		p.txnType = models.TransferTransaction
	}
}

func (p *transactionParser) finalizeMapping(isContact ContactVerifier, isAccount AccountVerifier) {
	processField := func(val string, isSource bool) {
		if val == "" {
			return
		}
		cleanVal := strings.ToLower(val)

		// 1. Match with Contact first
		if isContact(cleanVal) {
			p.txn.ContactName = cleanVal
			return
		}

		// 2. Match with Wallet (Account) name
		if isAccount(cleanVal) {
			if isSource {
				p.txn.SrcID = cleanVal
			} else {
				p.txn.DstID = cleanVal
			}
			return
		}

		// 3. Fallback to Remarks
		prefix := "to"
		if isSource {
			prefix = "from"
		}
		info := fmt.Sprintf("%s %s", prefix, val)

		if !strings.Contains(strings.ToLower(p.note), strings.ToLower(info)) {
			p.appendNote(info)
		}
	}

	processField(p.fromValue, true)
	processField(p.toValue, false)

	if p.txnType == models.TransferTransaction && p.txn.DstID == "" {
		// If it was supposed to be a transfer but no destination wallet found,
		// it's likely an expense or the user mentioned a contact.
		if p.txn.SrcID != "" && p.txn.ContactName != "" {
			// This could be a Debt transaction (Lend/Recover)
			// The subcategoryAIParser should have set the correct type and subcategory.
		} else {
			p.txnType = models.ExpenseTransaction
			if p.subcategory == "fin-transfer" {
				p.subcategory = "misc-misc"
			}
		}
	}

	if isDebtTransaction(p.subcategory) {
		if p.txn.ContactName == "" {
			var rawTarget string
			if p.toValue != "" && !isAccount(strings.ToLower(p.toValue)) {
				rawTarget = p.toValue
			} else if p.fromValue != "" && !isAccount(strings.ToLower(p.fromValue)) {
				rawTarget = p.fromValue
			}
			if rawTarget != "" {
				p.appendNote(fmt.Sprintf("[Person: %s]", rawTarget))
			}
		}
	}
}

func (p *transactionParser) isFormalSubcategoryID() bool {
	for _, subcat := range models.TxnSubcategories {
		if subcat.ID == p.subcategory {
			return true
		}
	}
	return false
}

func (p *transactionParser) appendNote(s string) {
	if p.note == "" {
		p.note = s
	} else {
		p.note += " " + s
	}
}

func (p *transactionParser) flushBuffer(key string, buffer []string, isContact ContactVerifier, isAccount AccountVerifier) {
	if len(buffer) == 0 {
		return
	}
	if key != "" {
		p.assignValue(key, strings.Join(buffer, " "), isAccount)
		return
	}

	// Bare wallet/contact mentions ("salary 50k ebl", "lent 500 karim") carry
	// routing info even without a preposition. Wallet names are stripped from
	// the classifier text (noise); contact names stay in it — the person is
	// context the classifier uses (matching keyed from/to behavior).
	kept := make([]string, 0, len(buffer))
	for _, w := range buffer {
		token := strings.ToLower(strings.Trim(w, ".,!?"))
		if isContact(token) {
			p.bareContacts = append(p.bareContacts, token)
		} else if isAccount(token) {
			p.bareAccounts = append(p.bareAccounts, token)
			continue
		}
		kept = append(kept, w)
	}
	if len(kept) == 0 {
		return
	}
	val := strings.Join(kept, " ")
	if p.subcategory != "" {
		p.subcategory += " " + val
	} else {
		p.subcategory = val
	}
}

func (p *transactionParser) assignValue(key, value string, isAccount AccountVerifier) {
	switch key {
	case "from":
		p.fromValue = value
	case "to":
		p.toValue = value
	case "in", "into":
		// "in"/"into" reads as a destination only when it points at a wallet
		// ("got salary in ebl"). Otherwise it's ordinary text ("lunch in office")
		// and goes back to the subcategory buffer so it isn't hijacked.
		if isAccount(strings.ToLower(value)) {
			p.toValue = value
			return
		}
		val := key + " " + value
		if p.subcategory != "" {
			p.subcategory += " " + val
		} else {
			p.subcategory = val
		}
	case "on":
		if isDateKeyword(strings.ToLower(value)) {
			p.date = value
			return
		}
		// Wallet check first so a wallet named like a date still wins.
		if isAccount(strings.ToLower(value)) {
			p.fromValue = value
			return
		}
		if _, err := pkg.ParseDate(value, p.tz); err == nil {
			p.date = value
			return
		}
		val := "on " + value
		if p.subcategory != "" {
			p.subcategory += " " + val
		} else {
			p.subcategory = val
		}
	case "at":
		p.time = value
	case "note":
		p.note = value
	}
}

func (p *transactionParser) cleanSubcategory() {
	p.subcategory = strings.TrimSpace(p.subcategory)
	lower := strings.ToLower(p.subcategory)
	if strings.HasPrefix(lower, "for ") {
		p.subcategory = p.subcategory[4:]
	}
	if strings.HasSuffix(lower, " for") {
		p.subcategory = p.subcategory[:len(p.subcategory)-4]
	}
}

func isStandardKeyword(w string) bool {
	switch w {
	case "from", "to", "in", "into", "on", "at":
		return true
	}
	return false
}

func isDateKeyword(w string) bool {
	switch w {
	case "yesterday", "today", "tomorrow":
		return true
	}
	// Full names only — short forms (fri, sat) collide too easily with
	// wallet/contact names, so those need an explicit "on".
	return pkg.IsFullWeekdayName(w)
}

func isDebtTransaction(subID string) bool {
	switch subID {
	case models.BorrowSubID, models.BorrowReturnSubID, models.LendSubID, models.LendRecoverySubID:
		return true
	}
	return false
}

func (p *transactionParser) ensureTypeMatchesCategory() {
	// List of Subcategories that are ALWAYS Income
	switch p.subcategory {
	case "fin-sal", "fin-prof", "fin-interest", "fin-borrow", "fin-recover", "fin-loan", "misc-init":
		p.txnType = models.IncomeTransaction
	// List of Subcategories that are ALWAYS Expense
	case "fin-repay", "fin-lend", "fin-return", "fin-tax", "fin-charge", "fin-ins":
		p.txnType = models.ExpenseTransaction
	// List of Subcategories that are ALWAYS Transfer (safety net over AI intent)
	case "fin-with", "fin-deposit", "fin-transfer":
		p.txnType = models.TransferTransaction
	}
}

func (p *transactionParser) parseTransaction() error {
	p.ensureTypeMatchesCategory()
	p.txn.Type = p.txnType
	p.txn.SubcategoryID = p.subcategory
	p.txn.Remarks = p.note
	p.applyBareMentions()
	p.setDefaultSourceDestination()

	if p.txn.SubcategoryID == "" {
		if p.txn.Type == models.TransferTransaction {
			if p.txn.SrcID == "cash" {
				p.txn.SubcategoryID = "fin-deposit"
			} else if p.txn.DstID == "cash" {
				p.txn.SubcategoryID = "fin-with"
			} else if p.txn.DstID == "credit" {
				p.txn.SubcategoryID = "fin-ccpay"
			}
		} else {
			p.txn.SubcategoryID = "misc-misc"
		}
	}
	if err := p.parseAmount(); err != nil {
		return err
	}
	return p.parseTransactionTime()
}

func (p *transactionParser) parseAmount() error {
	var err error
	p.txn.Amount, err = strconv.ParseFloat(p.amount, 64)
	return err
}

// applyBareMentions routes preposition-less wallet and contact mentions into
// the slots the final transaction type needs. Runs after type/subcategory are
// settled and never overrides an explicitly keyed value (from/to/in). The
// debt path attaches its person earlier; this covers the non-debt cases.
func (p *transactionParser) applyBareMentions() {
	if p.txn.ContactName == "" && len(p.bareContacts) > 0 {
		p.txn.ContactName = p.bareContacts[0]
	}
	for _, acc := range p.bareAccounts {
		switch p.txn.Type {
		case models.IncomeTransaction:
			if p.txn.DstID == "" {
				p.txn.DstID = acc
			}
		case models.ExpenseTransaction:
			if p.txn.SrcID == "" {
				p.txn.SrcID = acc
			}
		case models.TransferTransaction:
			switch {
			// "withdraw 5k ebl" takes from the bank; "deposit 5k ebl" puts into it.
			case p.txn.SubcategoryID == "fin-with" && p.txn.SrcID == "":
				p.txn.SrcID = acc
			case p.txn.SubcategoryID == "fin-deposit" && p.txn.DstID == "":
				p.txn.DstID = acc
			case p.txn.SrcID == "":
				p.txn.SrcID = acc
			case p.txn.DstID == "":
				p.txn.DstID = acc
			}
		}
	}
}

func (p *transactionParser) setDefaultSourceDestination() {
	if p.txn.Type == models.ExpenseTransaction || p.txn.Type == models.TransferTransaction {
		if p.txn.SrcID == "" {
			p.txn.SrcID = "cash"
		}
	}
	if p.txn.Type == models.IncomeTransaction || p.txn.Type == models.TransferTransaction {
		if p.txn.DstID == "" {
			p.txn.DstID = "cash"
		}
	}
}

func (p *transactionParser) parseTransactionTime() error {
	// ParseDate handles keywords, weekdays, ordinals and explicit formats;
	// empty date means today.
	date, err := pkg.ParseDate(p.date, p.tz)
	if err != nil {
		return err
	}
	year, month, day := date.Date()

	tim, err := pkg.ParseTime(p.time, p.tz)
	if err != nil {
		return err
	}

	// A date without a time means midnight of that day, not the current clock.
	var hour, minute, second int
	if p.date == "" || p.time != "" {
		hour, minute, second = tim.Clock()
	}
	p.txn.Timestamp = time.Date(year, month, day, hour, minute, second, 0, p.tz).Unix()
	return nil
}

func (p *transactionParser) isVerbKeyword(keyword string) bool {
	switch keyword {
	case "transfer", "transferred", "move", "moved", "send", "sent":
		p.txnType = models.TransferTransaction
		p.subcategory = "fin-transfer"
	case "withdraw", "withdrew", "cashout":
		p.txnType = models.TransferTransaction
		p.subcategory = "fin-with"
		p.toValue = "cash"
	case "deposit", "deposited", "cashin":
		p.txnType = models.TransferTransaction
		p.subcategory = "fin-deposit"
		p.fromValue = "cash"
	case "expense", "spend", "spent", "paid", "pay", "cost":
		p.txnType = models.ExpenseTransaction
	case "sell", "sold", "sale", "sales":
		p.txnType = models.IncomeTransaction
	case "giveaway", "donate", "donated", "gifted":
		p.txnType = models.ExpenseTransaction
		p.subcategory = "misc-gift"
	case "flexi", "recharge", "top-up":
		p.txnType = models.ExpenseTransaction
		p.subcategory = "fin-flexi"
	case "income", "earn", "earned", "received", "gained", "add", "plus":
		p.txnType = models.IncomeTransaction
		if keyword == "add" || keyword == "plus" {
			p.subcategory = "misc-init"
		}
	case "borrow", "borrowed":
		p.txnType = models.IncomeTransaction
		p.subcategory = models.BorrowSubID
	case "return", "returned", "repaid", "pay-back":
		p.txnType = models.ExpenseTransaction
		p.subcategory = models.BorrowReturnSubID
	case "lend", "lent":
		p.txnType = models.ExpenseTransaction
		p.subcategory = models.LendSubID
	case "recover", "recovered", "collect", "collected", "get-back":
		p.txnType = models.IncomeTransaction
		p.subcategory = models.LendRecoverySubID
	default:
		return false
	}
	return true
}
