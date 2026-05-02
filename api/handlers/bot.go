package handlers

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/modules/google"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"
	pkgtg "github.com/masudur-rahman/expense-tracker-bot/pkg/telegram"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"github.com/spf13/pflag"
	"gopkg.in/telebot.v3"
)

func StartTrackingExpenses(ctx telebot.Context) error {
	payload := ctx.Message().Payload
	logr.DefaultLogger.Infof("Start command received from %d with payload: %s", ctx.Sender().ID, payload)

	us := all.GetServices().User
	user, err := us.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		if !models.IsErrNotFound(err) {
			return ctx.Send(models.ErrCommonResponse(err))
		}
		user = &models.Profile{
			TelegramID: ctx.Sender().ID,
			Username:   ctx.Sender().Username,
			FirstName:  ctx.Sender().FirstName,
			LastName:   ctx.Sender().LastName,
		}
		if err = us.SignUp(user); err != nil {
			return ctx.Send(models.ErrCommonResponse(err))
		}
		ctx.Set("new-user", true)
	}

	if err = ensureDefaultWallet(user.ID); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	if strings.HasPrefix(payload, "login_") {
		sessionID := strings.TrimPrefix(payload, "login_")
		if ctx.Get("new-user") != nil {
			_ = sendStartText(ctx)
		}
		return HandleQRLogin(sessionID, ctx)
	}

	return sendStartText(ctx)
}

func ensureDefaultWallet(userID int64) error {
	_, err := all.GetServices().Wallet.GetWalletByShortName(userID, "cash")
	if err == nil {
		return nil
	}

	acc := &models.Wallet{
		UserID:    userID,
		Type:      models.CashAccount,
		ShortName: "cash",
		Name:      "Cash in Hand",
	}
	return all.GetServices().Wallet.CreateWallet(acc)
}

func sendStartText(ctx telebot.Context) error {
	firstName := html.EscapeString(ctx.Sender().FirstName)
	lastName := html.EscapeString(ctx.Sender().LastName)

	msg := fmt.Sprintf(`
<b>👋 Hi %s %s!!</b>

Welcome to <b>Expense Tracker Bot</b> 👛
Track <i>expenses, income, transfers, and loans</i> — right from chat.

✅ A <b>Cash</b> wallet is already created for you
  → You don't need to mention it unless you want another account
🏦 Add bank accounts anytime (BRAC, EBL, DBBL, etc)
👥 Add people you lend to or borrow from when needed

────────────────────

<b>➕ Add a Transaction</b>

You can do it <i>two ways</i>:

<b>1) Interactive (easy)</b>
Use the menu or send /newtxn and follow the steps.

<b>2) Just send a message (fast)</b>
Examples:

<code>lunch 250</code>
<code>groceries 1.5k</code>
<code>bought a new shirt for 2500</code>
<code>transfer 10k from brac to city</code>
<code>lent 5000 to karim</code>
<code>got bonus 20k</code>
<code>internet 500 on 1st</code>
<code>dinner 1500 yesterday</code>

💡 If no wallet is mentioned, <b>Cash is used by default</b>
⏱️ If no time is mentioned, <b>current time is used automatically</b>

────────────────────

<b>📊 Useful Commands</b>

/new      – Add accounts or people
/contacts – See contacts
/balance  – See balances
/summary  – Monthly summary
/report   – PDF report
/help     – Full guide & examples

────────────────────

🚀 Start by sending a transaction or tap the menu.
`, firstName, lastName)

	return ctx.Send(msg, telebot.ModeHTML)
}

func Help(ctx telebot.Context) error {
	return ctx.Send(fmt.Sprintf(`Click the following link to open the Usage documentation.
%s
`, "https://github.com/masudur-rahman/expense-tracker-bot/blob/main/README.md"))
}

func New(ctx telebot.Context) error {
	var callbackOpts CallbackOptions
	types := []CallbackType{ /*TransactionFlagTypeCallback,*/ AccountTypeCallback, UserTypeCallback}
	inlineButtons := make([]telebot.InlineButton, 0, 2)
	for _, typ := range types {
		callbackOpts.Type = typ
		btn := generateInlineButton(callbackOpts, typ)
		inlineButtons = append(inlineButtons, btn)
	}

	return ctx.Send("Select One", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	})
}

func ListContacts(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	contacts, err := all.GetServices().Contact.ListContacts(user.ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(printContacts(contacts), telebot.ModeMarkdown)
}

func printContacts(contacts []models.Contacts) string {
	if len(contacts) == 0 {
		return "No contacts found. Add one with /new"
	}
	var sb strings.Builder
	sb.WriteString("👥 *Your Contacts*\n")
	sb.WriteString(models.Separator + "\n")
	for _, c := range contacts {
		name := c.FullName
		if name == "" {
			name = c.NickName
		}
		sb.WriteString(fmt.Sprintf("👤 *%s* (`%s`)\n", name, c.NickName))
		if c.NetBalance > 0 {
			sb.WriteString(fmt.Sprintf("   ➕ `%.2f` _they owe you_\n\n", c.NetBalance))
		} else if c.NetBalance < 0 {
			sb.WriteString(fmt.Sprintf("   ➖ `%.2f` _you owe them_\n\n", -c.NetBalance))
		} else {
			sb.WriteString("   ✅ _settled_\n\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

func NewUser(ctx telebot.Context) error {
	// /newuser <id> <name> <email>
	ui := pkg.SplitString(ctx.Text(), ' ')
	if len(ui) < 3 {
		return ctx.Send("⚠️ Usage: /newuser <id> <name> <email>")
	}
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}
	if err := all.GetServices().Contact.CreateContact(&models.Contacts{
		UserID:   user.ID,
		NickName: ui[1],
		FullName: ui[2],
		Email: func() string {
			if len(ui) >= 4 {
				return ui[3]
			}
			return ""
		}(),
	}); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("✅ Contact *%s* added!", ui[2]), telebot.ModeMarkdown)
}

func AddAccount(ctx telebot.Context) error {
	// <type (Cash or Bank)> <unique-short-name> <Wallet Name>
	aci := pkg.SplitString(ctx.Text(), ' ')
	if len(aci) != 4 {
		return ctx.Send("⚠️ Usage: /new <type> <short-name> <wallet name>")
	}
	acc := &models.Wallet{
		Type:      models.WalletType(aci[1]),
		ShortName: aci[2],
		Name:      aci[3],
	}
	if err := all.GetServices().Wallet.CreateWallet(acc); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("✅ Wallet *%s* added!", aci[3]), telebot.ModeMarkdown)
}

func ListWallets(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	wallets, err := all.GetServices().Wallet.ListWallets(user.ID)
	if err != nil {
		return err
	}

	return ctx.Send(printWallets(wallets), telebot.ModeMarkdown)
}

func printWallets(wallets []models.Wallet) string {
	if len(wallets) == 0 {
		return "No wallets found. Add one with /new"
	}
	var sb strings.Builder
	sb.WriteString("💳 *Your Wallets*\n")
	sb.WriteString(models.Separator + "\n")
	var total float64
	for _, w := range wallets {
		icon := "💵"
		if w.Type == models.BankAccount {
			icon = "🏦"
		}
		sb.WriteString(fmt.Sprintf("%s *%s* (`%s`)\n", icon, w.Name, w.ShortName))
		sb.WriteString(fmt.Sprintf("   Balance: `%.2f`\n\n", w.Balance))
		total += w.Balance
	}
	sb.WriteString(models.Separator + "\n")
	sb.WriteString(fmt.Sprintf("💰 *Total: `%.2f`*", total))
	return sb.String()
}

/*
If users will be able to select options from the UI, it's ideal to design the input sequence in a way that guides them through the available options. Here's a suggested sequence that facilitates option selection:

1. Type: Ask the user to select the type of transaction (Expense, Income, Transfer). Present the available options as buttons or a dropdown menu.
2. Subcategory: Based on the selected type, present the relevant subcategories as options for the user to choose from. Display them as buttons or in a dropdown menu.
3. Amount: Once the subcategory is selected, prompt the user to enter the monetary amount of the transaction.
4. SrcID/DstID: Depending on the type of transaction, provide the appropriate options for the source ID (for Expense/Transfer) or destination ID (for Income/Transfer). This could be a dropdown menu or a list of selectable options.
5. Contacts (for Loan/Borrow): If the selected subcategory involves a person (Loan or Borrow), present the relevant users as options for the user to select from. Display them as buttons or in a dropdown menu.
6. Remarks: Provide an optional input field for the user to enter any additional remarks or notes related to the transaction.

By structuring the input sequence in this way, users can easily navigate through the available options and make their selections. It enhances the user experience by presenting a guided interface that reduces the chance of errors or confusion during the input process.
*/

type TransactionOptions struct {
	Type     string
	Amount   float64
	SubCatID string
	SrcID    string
	DstID    string
	UserID   string
	Remarks  string
}

func parseTransactionFlags(txnString string) (TransactionCallbackOptions, error) {
	var txnOpts TransactionCallbackOptions

	var typ string
	set := pflag.NewFlagSet("transaction", pflag.ContinueOnError)
	set.StringVarP(&typ, "type", "t", string(models.ExpenseTransaction), "Type of the transaction")
	set.StringVarP(&txnOpts.SubcategoryID, "subcat", "s", "misc-misc", "Subcategory for the transaction")
	set.StringVarP(&txnOpts.SrcID, "src", "f", "cash", "Source wallet for the transaction")
	set.StringVarP(&txnOpts.DstID, "dst", "d", "", "Destination wallet for the transaction")
	set.StringVarP(&txnOpts.ContactName, "contact", "u", "", "Contact associated with the loan/borrow")
	set.StringVarP(&txnOpts.Remarks, "remarks", "r", "", "Remarks for the transaction")
	txnOpts.Type = models.TransactionType(typ)

	args := pkg.SplitString(txnString, ' ')
	err := set.Parse(args)
	if err != nil {
		return TransactionCallbackOptions{}, err
	}

	if len(set.Args()) > 0 {
		_, err = fmt.Sscanf(set.Args()[0], "%f", &txnOpts.Amount)
	}
	txnOpts.NextStep = StepRemarks
	txnOpts.CategoryID = strings.Split(txnOpts.SubcategoryID, "-")[0]

	return txnOpts, err
}

/*
/txn <amount> -t=<type> -s=<subcat> -f=<src> -d=<dst> -u=<user> -r=<remarks>
*/

//func AddNewTransactions(ctx telebot.Context) error {
//	flags := strings.SplitN(ctx.Text(), " ", 2)
//	if len(flags) != 2 {
//		return ctx.Send("no argument provided for the transaction")
//	}
//
//	txnOpts, err := parseTransactionFlags(flags[1])
//	if err != nil {
//		return ctx.Send(err.Error())
//	}
//	params := models.Transaction{
//		Amount:        txnOpts.Amount,
//		SubcategoryID: txnOpts.SubCatID,
//		Type:          models.TransactionType(txnOpts.Type),
//		SrcID:         txnOpts.SrcID,
//		DstID:         txnOpts.DstID,
//		ContactName:        txnOpts.ContactName,
//		Timestamp:     time.Now().Unix(),
//		Remarks:       txnOpts.Remarks,
//	}
//	err = all.GetServices().Txn.AddTransaction(params)
//	if err != nil {
//		return ctx.Send(err.Error())
//	}
//
//	return ctx.Send("Transaction added")
//}

func ListTransactions(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	// Fetch transactions from the last 30 days only
	startTime := time.Now().AddDate(0, 0, -30).Unix()
	txns, err := all.GetServices().Txn.ListTransactionsByTime(user.ID, "", startTime, time.Now().Unix())
	if err != nil {
		return err
	}

	if len(txns) == 0 {
		return ctx.Send("No transactions found in the last 30 days.")
	}

	pageSize := 10
	page := 1

	formatted := pkgtg.FormatTransactionList(txns, page, pageSize)
	callbackOpts := CallbackOptions{
		Type: ListPaginationTypeCallback,
		Pagination: PaginationCallbackOptions{
			Page: page,
		},
	}

	return ctx.Send(formatted, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generatePaginationKeyboard(callbackOpts, page, len(txns) > pageSize),
		},
	})
}

func ListExpenses(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	txns, err := all.GetServices().Txn.ListTransactionsByTime(user.ID, models.ExpenseTransaction, 0, time.Now().Unix())
	if err != nil {
		return err
	}

	if len(txns) == 0 {
		return ctx.Send("No expenses found.")
	}

	pageSize := 10
	page := 1

	formatted := pkgtg.FormatTransactionList(txns, page, pageSize)
	callbackOpts := CallbackOptions{
		Type: ListPaginationTypeCallback,
		Pagination: PaginationCallbackOptions{
			Page:   page,
			IsExps: true,
		},
	}

	return ctx.Send(formatted, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generatePaginationKeyboard(callbackOpts, page, len(txns) > pageSize),
		},
	})
}

func SyncSQLiteDatabase(ctx telebot.Context) error {
	db := configs.TrackerConfig.Database
	if !(db.Type == configs.DatabaseSQLite && db.SQLite.SyncToDrive) {
		return ctx.Send("⚠️ SQLite with Drive sync must be enabled for this.")
	}

	if err := google.SyncDatabaseToDrive(); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send("✅ Database synced to Google Drive")
}
