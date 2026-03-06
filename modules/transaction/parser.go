package transaction

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/modules/ai"
	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"
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
	verbFound   bool
}

func ParseTransaction(text string, isContact ContactVerifier, isAccount AccountVerifier) (models.Transaction, error) {
	// --- STEP 0: Safety Defaults ---
	if isContact == nil {
		isContact = func(name string) bool { return false }
	}
	if isAccount == nil {
		isAccount = func(name string) bool { return false }
	}

	p := transactionParser{}

	// --- STEP 1: Find Amount ---
	reAmount := regexp.MustCompile(`(?i)(?:(?:total|tk|taka|amount)\s*)?(\d+(?:\.\d+)?)\s*(k)?(?:\s*(?:tk|taka|bdt))?`)
	loc := reAmount.FindStringSubmatchIndex(text)
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

	// --- STEP 2: Tokenize & Scan ---
	words := strings.Fields(textWithoutAmount)
	p.txnType = models.ExpenseTransaction

	var currentKey string
	var currentBuffer []string

	for i := 0; i < len(words); i++ {
		word := words[i]
		lowerWord := strings.ToLower(word)

		if lowerWord == "note" {
			p.flushBuffer(currentKey, currentBuffer, isAccount)
			if i+1 < len(words) {
				p.note = strings.Join(words[i+1:], " ")
			}
			currentKey = ""
			currentBuffer = nil
			break
		}
		if p.isVerbKeyword(lowerWord) {
			p.verbFound = true
			p.flushBuffer(currentKey, currentBuffer, isAccount)
			currentBuffer = []string{}
			currentKey = ""
			continue
		}
		if isDateKeyword(lowerWord) {
			p.flushBuffer(currentKey, currentBuffer, isAccount)
			p.date = lowerWord
			currentBuffer = []string{}
			currentKey = ""
			continue
		}
		if isStandardKeyword(lowerWord) {
			p.flushBuffer(currentKey, currentBuffer, isAccount)
			currentKey = lowerWord
			currentBuffer = []string{}
		} else {
			currentBuffer = append(currentBuffer, word)
		}
	}
	p.flushBuffer(currentKey, currentBuffer, isAccount)
	p.cleanSubcategory()

	// --- STEP 3: Enrich Context (Pre-AI) ---
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

	p.subcategory = strings.ToLower(p.subcategory)
	if subcat, exist := cache.GetCache(p.subcategory); exist {
		p.subcategory = subcat
		return nil
	}
	inputText := p.subcategory
	subcat, err := ai.TxnCategoryGenerator(context.Background(), inputText)
	if err != nil {
		return err
	}

	_ = cache.SetCache(inputText, subcat, -1)
	if dbErr := configs.InsertAICache(models.AICache{
		InputText:     inputText,
		SubcategoryID: subcat,
		CreatedAt:     time.Now().Unix(),
	}); dbErr != nil {
		logr.DefaultLogger.Errorw("Failed to persist AI cache", "error", dbErr.Error())
	}
	p.subcategory = subcat
	return nil
}

func (p *transactionParser) finalizeMapping(isContact ContactVerifier, isAccount AccountVerifier) {
	processField := func(val string, isSource bool) {
		if val == "" {
			return
		}
		cleanVal := strings.ToLower(val)

		if isAccount(cleanVal) {
			if isSource {
				p.txn.SrcID = cleanVal
			} else {
				p.txn.DstID = cleanVal
			}
			return
		}

		if isDebtTransaction(p.subcategory) {
			if isContact(cleanVal) {
				p.txn.ContactName = cleanVal
				return
			}
			return
		}

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
		p.txnType = models.ExpenseTransaction
		if p.subcategory == "fin-transfer" {
			p.subcategory = "misc-misc"
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

func (p *transactionParser) flushBuffer(key string, buffer []string, isAccount AccountVerifier) {
	if len(buffer) == 0 {
		return
	}
	val := strings.Join(buffer, " ")
	if key != "" {
		p.assignValue(key, val, isAccount)
	} else {
		if p.subcategory != "" {
			p.subcategory += " " + val
		} else {
			p.subcategory = val
		}
	}
}

func (p *transactionParser) assignValue(key, value string, isAccount AccountVerifier) {
	switch key {
	case "from":
		p.fromValue = value
	case "to":
		p.toValue = value
	case "on":
		if isDateKeyword(value) {
			p.date = value
			return
		}
		if _, err := pkg.ParseDate(value); err == nil {
			p.date = value
			return
		}
		if isAccount(strings.ToLower(value)) {
			p.fromValue = value
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
	case "from", "to", "on", "at":
		return true
	}
	return false
}

func isDateKeyword(w string) bool {
	switch w {
	case "yesterday", "today", "tomorrow":
		return true
	}
	return false
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
	case "fin-sal", "fin-prof", "fin-interest", "fin-borrow", "fin-recover":
		p.txnType = models.IncomeTransaction
	// List of Subcategories that are ALWAYS Expense
	case "fin-repay", "fin-lend", "fin-return":
		p.txnType = models.ExpenseTransaction
	}
}

func (p *transactionParser) parseTransaction() error {
	p.ensureTypeMatchesCategory()
	p.txn.Type = p.txnType
	p.txn.SubcategoryID = p.subcategory
	p.txn.Remarks = p.note
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
	var year, day, hour, minute, second int
	var month time.Month

	if isDateKeyword(strings.ToLower(p.date)) {
		now := time.Now()
		switch strings.ToLower(p.date) {
		case "yesterday":
			now = now.AddDate(0, 0, -1)
		case "tomorrow":
			now = now.AddDate(0, 0, 1)
		}
		year, month, day = now.Date()
	} else {
		date, err := pkg.ParseDate(p.date)
		if err != nil {
			return err
		}
		year, month, day = date.Date()
	}
	tim, err := pkg.ParseTime(p.time)
	if err != nil {
		return err
	}

	if p.date != "" && p.time == "" {
		hour, minute, second = 0, 0, 0
	} else {
		hour, minute, second = tim.Clock()
	}
	p.txn.Timestamp = time.Date(year, month, day, hour, minute, second, 0, time.Local).Unix()
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
	case "income", "earn", "earned", "received", "gained":
		p.txnType = models.IncomeTransaction
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
