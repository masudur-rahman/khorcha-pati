package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/models/gqtypes"
	"github.com/masudur-rahman/expense-tracker-bot/modules/convert"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"
	"github.com/masudur-rahman/expense-tracker-bot/templates"

	"gopkg.in/telebot.v3"
)

type ReportCallbackOptions struct {
	Duration SummaryDuration `json:"duration"`
}

func TransactionReportCallback(ctx telebot.Context) error {
	callbackOpts := CallbackOptions{
		Type: ReportTypeCallback,
	}

	inlineButtons := generateReportDurationInlineButton(callbackOpts)

	return ctx.Send("Select Duration for the Report", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	})
}

func generateReportDurationInlineButton(callbackOpts CallbackOptions) []telebot.InlineButton {
	durations := []struct {
		val   SummaryDuration
		label string
	}{
		{DurationOneWeek, "1 Week"},
		{DurationThisMonth, "This Month"},
		{DurationOneMonth, "One Month"},
		{DurationHalfYear, "6 Months"},
		{DurationThisYear, "This Year"},
		{DurationOneYear, "1 Year"},
	}
	inlineButtons := make([]telebot.InlineButton, 0, len(durations))
	for _, d := range durations {
		callbackOpts.Report.Duration = d.val
		btn := generateInlineButton(callbackOpts, d.label)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons
}

func handleReportCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	report, err := generateReport(ctx, callbackOpts.Report)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	pdfFile, err := GenerateTransactionStatementFromTemplate(report, "")
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}
	defer os.Remove(pdfFile)

	return ctx.Send(&telebot.Document{
		File:     telebot.FromDisk(pdfFile),
		FileName: "transaction_report.pdf",
	})
}

func generateSampleJSONReport(report gqtypes.Report) error { //nolint:unused // kept for local debugging
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pkg.ProjectDirectory+"/templates/"+"sample_report.json", data, 0644)
}

func generateReport(ctx telebot.Context, rop ReportCallbackOptions) (gqtypes.Report, error) {
	now, startTime := time.Now(), CalculateStartTime(rop.Duration)

	svc := all.GetServices()
	user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return gqtypes.Report{}, err
	}

	txns, err := svc.Txn.ListTransactionsByTime(user.ID, "", startTime.Unix(), now.Unix())
	if err != nil {
		return gqtypes.Report{}, err
	}

	startTime = clampStartTime(startTime, user.CreatedAt, txns)

	report := gqtypes.Report{
		Name:      fmt.Sprintf("%v %v", user.FirstName, user.LastName),
		StartDate: startTime,
		EndDate:   now,
	}

	wallets, err := svc.Wallet.ListWallets(user.ID)
	if err == nil {
		report.Wallets = make([]gqtypes.Wallet, 0, len(wallets))
		for _, w := range wallets {
			report.Wallets = append(report.Wallets, convert.ToWalletAPIFormat(w))
		}
	}

	contacts, err := svc.Contact.ListContacts(user.ID)
	if err == nil {
		report.Contacts = make([]gqtypes.Contact, 0, len(contacts))
		for _, c := range contacts {
			report.Contacts = append(report.Contacts, convert.ToContactAPIFormat(c))
		}
	}

	txnApis := make([]gqtypes.Transaction, 0, len(txns))
	for _, txn := range txns {
		txnApis = append(txnApis, convert.ToTransactionAPIFormat(txn))
	}
	report.Transactions = txnApis
	FinalizeReportTxns(&report, now)

	summary, err := BuildSummary(svc, txns)
	if err != nil {
		return gqtypes.Report{}, err
	}
	report.Summary = summary

	report.TypeSummary = gqtypes.SortMapToSlice(summary.Type)
	report.CategorySummary, err = BuildTypeSeparatedSummary(svc, txns, CategoryKeyFn, svc.Txn.GetTxnCategoryName)
	if err != nil {
		return gqtypes.Report{}, err
	}
	report.SubcategorySummary, err = BuildTypeSeparatedSummary(svc, txns, SubcategoryKeyFn, svc.Txn.GetTxnSubcategoryName)
	if err != nil {
		return gqtypes.Report{}, err
	}

	report.TotalAmount, report.NetBalance = ComputeTotals(txns)

	return report, nil
}

func BuildSummary(svc *all.Services, txns []models.Transaction) (gqtypes.SummaryGroups, error) {
	summary := gqtypes.SummaryGroups{
		Type:        map[string]gqtypes.FieldCost{},
		Category:    map[string]gqtypes.FieldCost{},
		Subcategory: map[string]gqtypes.FieldCost{},
	}

	for _, txn := range txns {
		fc := summary.Type[string(txn.Type)]
		fc.Amount += txn.Amount
		summary.Type[string(txn.Type)] = fc

		fc = summary.Subcategory[txn.SubcategoryID]
		fc.Amount += txn.Amount
		summary.Subcategory[txn.SubcategoryID] = fc

		cat := strings.Split(txn.SubcategoryID, "-")[0]
		fc = summary.Category[cat]
		fc.Amount += txn.Amount
		summary.Category[cat] = fc
	}

	for k, fc := range summary.Type {
		fc.Name = k
		summary.Type[k] = fc
	}

	var err error
	for k, fc := range summary.Category {
		fc.Name, err = svc.Txn.GetTxnCategoryName(k)
		if err != nil {
			fc.Name = k // fallback to ID
		}
		summary.Category[k] = fc
	}

	for k, fc := range summary.Subcategory {
		fc.Name, err = svc.Txn.GetTxnSubcategoryName(k)
		if err != nil {
			fc.Name = k // fallback to ID
		}
		summary.Subcategory[k] = fc
	}

	return summary, nil
}

func CategoryKeyFn(txn models.Transaction) string {
	return strings.Split(txn.SubcategoryID, "-")[0]
}

func SubcategoryKeyFn(txn models.Transaction) string {
	return txn.SubcategoryID
}

// BuildTypeSeparatedSummary aggregates by key+type, resolves names, and returns sorted slice.
func BuildTypeSeparatedSummary(
	svc *all.Services,
	txns []models.Transaction,
	keyFn func(models.Transaction) string,
	nameFn func(string) (string, error),
) ([]gqtypes.FieldCost, error) {
	type compositeKey struct {
		key     string
		txnType string
	}

	agg := map[compositeKey]float64{}
	for _, txn := range txns {
		k := compositeKey{key: keyFn(txn), txnType: string(txn.Type)}
		agg[k] += txn.Amount
	}

	result := make([]gqtypes.FieldCost, 0, len(agg))
	for ck, amount := range agg {
		name, err := nameFn(ck.key)
		if err != nil {
			name = ck.key // fallback to ID
		}
		result = append(result, gqtypes.FieldCost{
			Name:   name,
			Amount: amount,
			Type:   ck.txnType,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Amount > result[j].Amount
	})

	return result, nil
}

// FinalizeReportTxns sorts transactions by date asc, fills RunningBalance, and stamps GeneratedAt.
func FinalizeReportTxns(report *gqtypes.Report, now time.Time) {
	sort.Slice(report.Transactions, func(i, j int) bool {
		return report.Transactions[i].Date.Before(report.Transactions[j].Date)
	})
	var running float64
	for i := range report.Transactions {
		switch report.Transactions[i].Type {
		case string(models.IncomeTransaction):
			running += report.Transactions[i].Amount
		case string(models.ExpenseTransaction):
			running -= report.Transactions[i].Amount
		}
		report.Transactions[i].RunningBalance = running
	}
	if now.IsZero() {
		now = time.Now()
	}
	report.GeneratedAt = now
}

func ComputeTotals(txns []models.Transaction) (totalAmount, netBalance float64) {
	for _, txn := range txns {
		totalAmount += txn.Amount
		switch txn.Type {
		case models.IncomeTransaction:
			netBalance += txn.Amount
		case models.ExpenseTransaction:
			netBalance -= txn.Amount
		}
	}
	return totalAmount, netBalance
}

func CalculateStartTime(duration SummaryDuration) time.Time {
	now, startTime := time.Now(), pkg.StartOfMonth()
	switch duration {
	case DurationOneWeek:
		startTime = now.AddDate(0, 0, -7)
	case DurationThisMonth:
		startTime = pkg.StartOfMonth()
	case DurationOneMonth:
		startTime = now.AddDate(0, -1, 0)
	case DurationHalfYear:
		startTime = now.AddDate(0, -6, 0)
	case DurationThisYear:
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	case DurationOneYear:
		startTime = now.AddDate(-1, 0, 0)
	case DurationAllTime:
		startTime = time.Time{}
	}
	return startTime
}

func GenerateTransactionStatementFromTemplate(report gqtypes.Report, title string) (string, error) {
	converter := string(configs.TrackerConfig.System.PDFGenerator)

	funcMap := template.FuncMap{
		"formatBDT":       FormatBDT,
		"toLower":         strings.ToLower,
		"balanceClass":    BalanceClass,
		"formatAmount":    FormatAmount,
		"formatDate":      FormatStatementDate,
		"formatDateRange": FormatStatementDateRange,
		"formatGenAt":     FormatGeneratedAt,
		"amountColor":     AmountColor,
		"typeBg":          TypeBgColor,
		"typeText":        TypeTextColor,
		"abs":             AbsFloat,
	}

	bodyBytes, err := executeTemplate("transaction_report.tmpl", funcMap, &report)
	if err != nil {
		return "", err
	}

	headerTmplName, footerTmplName := selectTemplateNames(converter)

	headerBytes, err := executeTemplate(headerTmplName, funcMap, &report)
	if err != nil {
		return "", err
	}

	footerBytes, err := executeTemplate(footerTmplName, funcMap, &report)
	if err != nil {
		return "", err
	}

	pdfFile, err := os.CreateTemp("", "transaction_report_*.pdf")
	if err != nil {
		return "", err
	}
	pdfFile.Close()

	if err = pkg.ConvertHTMLToPDF(converter, pdfFile.Name(), bodyBytes, headerBytes, footerBytes, title); err != nil {
		os.Remove(pdfFile.Name())
		return "", err
	}

	return pdfFile.Name(), nil
}

// selectTemplateNames returns header and footer template names for the converter.
func selectTemplateNames(converter string) (header, footer string) {
	if converter == string(configs.PDFGeneratorChromeDP) {
		return "header_cdp.tmpl", "footer_cdp.tmpl"
	}
	return "header.tmpl", "footer.tmpl"
}

// executeTemplate reads and executes a named template from the embedded FS.
func executeTemplate(name string, funcMap template.FuncMap, data any) ([]byte, error) {
	raw, err := templates.FS.ReadFile(name)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func FormatBDT(amount float64) string {
	if amount < 0 {
		return "-" + FormatBDT(-amount)
	}

	// Split into integer and decimal parts
	integerPart := int64(amount)
	decimalPart := amount - float64(integerPart)

	// Format integer part with BDT commas
	formattedInteger := formatBangladeshiCommas(integerPart)

	// Handle decimal part
	if decimalPart == 0 {
		return formattedInteger
	}

	// Format decimal to 2 places and trim leading zero
	decimalStr := strings.TrimPrefix(strconv.FormatFloat(decimalPart, 'f', 2, 64), "0")
	return formattedInteger + decimalStr
}

func formatBangladeshiCommas(n int64) string {
	str := strconv.FormatInt(n, 10)
	length := len(str)

	if length <= 3 {
		return str
	}

	// Reverse string for easier grouping
	reversed := reverseString(str)
	groups := []string{}

	// First group of 3 digits
	groups = append(groups, reversed[:3])
	reversed = reversed[3:]

	// Subsequent groups of 2 digits
	for len(reversed) > 0 {
		chunkSize := 2
		if len(reversed) < chunkSize {
			chunkSize = len(reversed)
		}
		groups = append(groups, reversed[:chunkSize])
		reversed = reversed[chunkSize:]
	}

	// Reverse groups and create formatted string
	formattedGroups := make([]string, len(groups))
	for i, group := range groups {
		formattedGroups[len(groups)-1-i] = reverseString(group)
	}

	return strings.Join(formattedGroups, ",")
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func BalanceClass(amount float64) string {
	if amount > 0 {
		return "income"
	}
	if amount < 0 {
		return "expense"
	}
	return ""
}

// FormatAmount renders a value as `৳1,234.56` to mirror the React Statement page.
func FormatAmount(amount float64) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	intPart := int64(amount)
	frac := int64((amount-float64(intPart))*100 + 0.5)
	if frac >= 100 {
		intPart++
		frac -= 100
	}
	intStr := strconv.FormatInt(intPart, 10)
	var b strings.Builder
	for i, r := range intStr {
		remaining := len(intStr) - i
		if i > 0 && remaining%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	out := fmt.Sprintf("৳%s.%02d", b.String(), frac)
	if neg {
		out = "-" + out
	}
	return out
}

// FormatStatementDate renders dd/mm to match the React table's date column.
func FormatStatementDate(t time.Time) string {
	return t.Format("02/01")
}

// FormatStatementDateRange renders "Month D, YYYY — Month D, YYYY".
func FormatStatementDateRange(start, end time.Time) string {
	const layout = "January 2, 2006"
	return fmt.Sprintf("%s — %s", start.Format(layout), end.Format(layout))
}

// FormatGeneratedAt renders the footer timestamp.
func FormatGeneratedAt(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	return t.Format("Jan 2, 2006, 3:04 PM")
}

const (
	colorIncome   = "#00875A"
	colorExpense  = "#DE350B"
	colorTransfer = "#0052CC"
	bgIncome      = "#E3FCEF"
	bgExpense     = "#FFEBE6"
	bgTransfer    = "#DEEBFF"
	colorMuted    = "#505F79"
	bgMuted       = "#F4F5F7"
)

// AmountColor returns the hex color that mirrors Statement.tsx's amountColor map.
func AmountColor(txnType string) string {
	switch txnType {
	case string(models.IncomeTransaction):
		return colorIncome
	case string(models.ExpenseTransaction):
		return colorExpense
	case string(models.TransferTransaction):
		return colorTransfer
	}
	return colorMuted
}

// TypeBgColor returns the badge background hex for a transaction type.
func TypeBgColor(txnType string) string {
	switch txnType {
	case string(models.IncomeTransaction):
		return bgIncome
	case string(models.ExpenseTransaction):
		return bgExpense
	case string(models.TransferTransaction):
		return bgTransfer
	}
	return bgMuted
}

// TypeTextColor returns the badge text hex for a transaction type.
func TypeTextColor(txnType string) string {
	return AmountColor(txnType)
}

// AbsFloat returns |x| for use in templates.
func AbsFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// clampStartTime returns the effective start date for a report period.
// Fallback chain: registration time → earliest txn time → startTime (computed).
func clampStartTime(startTime time.Time, registeredAt int64, txns []models.Transaction) time.Time {
	if registeredAt != 0 {
		if reg := time.Unix(registeredAt, 0); reg.After(startTime) {
			return reg
		}
		return startTime
	}
	// fallback: earliest transaction timestamp
	if len(txns) > 0 {
		earliest := txns[0].Timestamp
		for _, t := range txns[1:] {
			if t.Timestamp < earliest {
				earliest = t.Timestamp
			}
		}
		if first := time.Unix(earliest, 0); first.After(startTime) {
			return first
		}
	}
	return startTime
}
