package handlers

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/modules/google"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/pflag"
	"gopkg.in/telebot.v3"
)

func StartTrackingExpenses(ctx telebot.Context) error {
	us := all.GetServices().User
	user, err := us.GetUserByTelegramID(ctx.Sender().ID)
	if err == nil {
		return sendStartText(ctx)
	}
	if !models.IsErrNotFound(err) {
		logr.DefaultLogger.Errorw("Start error", "error", err.Error())
		return ctx.Send(err.Error())
	}
	user = &models.User{
		TelegramID: ctx.Sender().ID,
		Username:   ctx.Sender().Username,
		FirstName:  ctx.Sender().FirstName,
		LastName:   ctx.Sender().LastName,
	}
	if err = us.SignUp(user); err != nil {
		logr.DefaultLogger.Errorw("Sign up error", "error", err.Error())
		return ctx.Send(err.Error())
	}

	acc := &models.Account{
		UserID:    user.ID,
		Type:      models.CashAccount,
		ShortName: "cash",
		Name:      "Cash in Hand",
	}

	if err = all.GetServices().Account.CreateAccount(acc); err != nil {
		logr.DefaultLogger.Errorw("Create default cash account error", "error", err.Error())
		return ctx.Send(err.Error())
	}

	ctx.Set("new-user", true)
	return sendStartText(ctx)
}

func sendStartText(ctx telebot.Context) error {
	firstName := html.EscapeString(ctx.Sender().FirstName)
	lastName := html.EscapeString(ctx.Sender().LastName)

	msg := fmt.Sprintf(`
<b>👋 Hi %s %s!!</b>

Welcome to <b>Expense Tracker Bot</b> 👛
Track <i>expenses, income, transfers, and loans</i> — right from chat.

✅ A <b>Cash</b> account is already created for you
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

<code>spent 500 for lunch</code>
<code>spent 1000 for "Chicken Biriyani" from brac</code>
<code>borrow 2000 from john to brac</code>

💡 If no account is mentioned, <b>Cash is used by default</b>
⏱️ If no time is mentioned, <b>current time is used automatically</b>

Tip: wrap multi-word names in quotes
<i>"Chicken Biriyani"</i>, <i>"Bill payment"</i>

────────────────────

<b>📊 Useful Commands</b>

/new      – Add accounts or people
/balance  – See balances
/summary  – Monthly summary
/report   – PDF report
/help     – Full guide & examples

────────────────────

🚀 Start by sending a transaction or tap the menu.
`, firstName, lastName)

	return ctx.Send(msg, telebot.ModeHTML)
}

func Welcome(ctx telebot.Context) error {
	return ctx.Send(fmt.Sprintf(`Hello %v %v!
Welcome to Expense Tracker !
`, ctx.Sender().FirstName, ctx.Sender().LastName))
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

func ListUsers(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	users, err := all.GetServices().DebtorCreditor.ListDebtorCreditors(user.ID)
	if err != nil {
		return ctx.Send(err.Error())
	}

	return ctx.Send(pkg.FormatDocuments(users, "NickName", "FullName", "Balance"))
}

func NewUser(ctx telebot.Context) error {
	// /newuser <id> <name> <email>
	ui := pkg.SplitString(ctx.Text(), ' ')
	if len(ui) < 3 {
		return ctx.Send(`
Syntax unknown.
Format /newuser <id> <name> <email>
`)
	}
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(err.Error())
	}
	if err := all.GetServices().DebtorCreditor.CreateDebtorCreditor(&models.DebtorsCreditors{
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
		log.Println(err)
		return ctx.Send(err.Error())
	}

	return ctx.Send("New DebtorsCreditors added!")
}

func AddAccount(ctx telebot.Context) error {
	// <type (Cash or Bank)> <unique-short-name> <Account Name>
	aci := pkg.SplitString(ctx.Text(), ' ')
	if len(aci) != 4 {
		return ctx.Send(`
Syntax unknown.
Format /new <type> <unique-name> <Account Name>
`)
	}
	acc := &models.Account{
		Type:      models.AccountType(aci[1]),
		ShortName: aci[2],
		Name:      aci[3],
	}
	if err := all.GetServices().Account.CreateAccount(acc); err != nil {
		log.Println(err)
		return ctx.Send(err.Error())
	}

	return ctx.Send("New Account Added !")
}

func ListAccounts(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	accounts, err := all.GetServices().Account.ListAccounts(user.ID)
	if err != nil {
		return err
	}

	return ctx.Send(printAccounts(accounts))
}

func printAccounts(accounts []models.Account) string {
	buf := bytes.Buffer{}
	w := tabwriter.NewWriter(&buf, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, "ID\tType\tName\tBalance")
	for _, ac := range accounts {
		fmt.Fprintf(w, "%v\t%v\t%v\t%.2f\n", ac.ShortName, ac.Type, ac.Name, ac.Balance)
	}
	_ = w.Flush()
	return buf.String()
}

/*
If users will be able to select options from the UI, it's ideal to design the input sequence in a way that guides them through the available options. Here's a suggested sequence that facilitates option selection:

1. Type: Ask the user to select the type of transaction (Expense, Income, Transfer). Present the available options as buttons or a dropdown menu.
2. Subcategory: Based on the selected type, present the relevant subcategories as options for the user to choose from. Display them as buttons or in a dropdown menu.
3. Amount: Once the subcategory is selected, prompt the user to enter the monetary amount of the transaction.
4. SrcID/DstID: Depending on the type of transaction, provide the appropriate options for the source ID (for Expense/Transfer) or destination ID (for Income/Transfer). This could be a dropdown menu or a list of selectable options.
5. DebtorsCreditors (for Loan/Borrow): If the selected subcategory involves a person (Loan or Borrow), present the relevant users as options for the user to select from. Display them as buttons or in a dropdown menu.
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
	set.StringVarP(&txnOpts.SrcID, "src", "f", "cash", "Source account for the transaction")
	set.StringVarP(&txnOpts.DstID, "dst", "d", "", "Destination account for the transaction")
	set.StringVarP(&txnOpts.DebtorCreditorName, "user", "u", "", "DebtorsCreditors associated with the loan/borrow")
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
//		DebtorCreditorName:        txnOpts.DebtorCreditorName,
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

	txns, err := all.GetServices().Txn.ListTransactions(user.ID)
	if err != nil {
		return err
	}

	printer := configs.GetDefaultPrinter()
	//printer.WithRenderType(pkg.RenderTypeMarkdown)
	printer.WithStyle(table.StyleLight)
	printer.WithExceptColumns([]string{"ID"})
	defer printer.ClearColumns()
	printer.PrintDocuments(txns)

	return ctx.Send(pkg.FormatDocuments(txns, "Timestamp", "Amount", "Type"))
}

func ListExpenses(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	txns, err := all.GetServices().Txn.ListTransactionsByTime(user.ID, models.ExpenseTransaction, pkg.StartOfMonth().Unix(), time.Now().Unix())
	if err != nil {
		return err
	}

	printer := configs.GetDefaultPrinter()
	//printer.WithRenderType(pkg.RenderTypeMarkdown)
	printer.WithStyle(table.StyleLight)
	printer.WithExceptColumns([]string{"ID"})
	defer printer.ClearColumns()
	printer.PrintDocuments(txns)

	return ctx.Send(pkg.FormatDocuments(txns, "Timestamp", "Amount", "Type"))
}

func SyncSQLiteDatabase(ctx telebot.Context) error {
	db := configs.TrackerConfig.Database
	if !(db.Type == configs.DatabaseSQLite && db.SQLite.SyncToDrive) {
		return ctx.Send("Database needs to be SQLite and sync needs to be enabled")
	}

	if err := google.SyncDatabaseToDrive(); err != nil {
		return ctx.Send(fmt.Sprintf("Database sync failed, reason: %v", err))
	}

	return ctx.Send("Database synced to google drive successfully")
}
