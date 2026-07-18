package transaction

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
	"github.com/masudur-rahman/khorcha-pati/pkg"
)

var testInputs = []string{
	// The "Minimalist"
	"rickshaw 50",
	"lunch 250",
	"groceries 1.5k",
	"haircut 200",
	"gym 1000",
	"electricity bill 1200",
	"internet 500",
	"bus fare 30",
	"netflix 1200",
	"tea 20",

	// The "Context Provider"
	"bought a new shirt for eid 2500",
	"sent 5000 to ammu for medicine",
	"paid 1500 for wifi bill",
	"kfc dinner with team 1200",
	"bought gift for rahim wedding 2000",
	"repair bike brake shoe 450",
	"advance payment for house rent 10k",
	"donation for flood victims 500",
	"snacks for office party 800",
	"medicine for fever and cough 350",

	// The "Financial Manager"
	"transfer 10k from brac to city",
	"cashout 5000 from bkash",
	"send 2000 to driver on nagad",
	"deposit 50k to dbbl",
	"withdraw 10000 from atm",
	"credit card bill payment 15500",
	"load 500 to my gp number",
	"transfer 5000 to savings",
	"received salary 65k",
	"profit from stocks 2.5k",

	// The "Lender/Borrower"
	"lent 5000 to karim",
	"borrowed 2000 from rifat",
	"returned 1000 to rifat",
	"recovered 2500 from karim",
	"gave 500 loan to security guard",
	"took 1000 from ammu",
	"paid back 500 to shopkeeper",
	"lent 10k to office colleague",
	"borrow 500 for emergency",
	"collected 2000 from batchmate",

	// The "Natural Speaker"
	"spent 500 yesterday for pizza",
	"sold old chair 1200",
	"got bonus 20k",
	"lost wallet with 500 taka",
	"found 100 taka on road",
	"sold my bike for 50k",
	"2000 taka given to maid",
	"shopping 5k from bashundhara city",
	"total 450 cost for uber",
	"received 1500 from tuition",
}

func initCache() {
	cfg := cache.Config{
		Type: cache.CacheMap,
	}
	cache.Init(cfg)
}

// TestParseTransaction_destinationPreposition verifies that "in"/"into" read as a
// destination when they point at a wallet, but stay as plain text otherwise.
func TestParseTransaction_destinationPreposition(t *testing.T) {
	initCache()
	// Seed the cache so free-text phrases resolve deterministically without the AI.
	_ = cache.SetCache("got salary", `{"intent":"income","subcategory_id":"fin-sal"}`, -1)
	_ = cache.SetCache("lunch in office", `{"intent":"expense","subcategory_id":"food-rest"}`, -1)

	contacts := func(string) bool { return false }
	accounts := func(name string) bool {
		switch strings.ToLower(name) {
		case "cash", "ebl", "dbbl":
			return true
		}
		return false
	}

	tests := []struct {
		name    string
		text    string
		wantTyp models.TransactionType
		wantSrc string
		wantDst string
	}{
		{"income in wallet", "got salary 65k in ebl", models.IncomeTransaction, "", "ebl"},
		{"transfer into wallet", "transfer 500 from cash into dbbl", models.TransferTransaction, "cash", "dbbl"},
		{"in non-wallet stays text", "lunch 250 in office", models.ExpenseTransaction, "cash", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransaction(tt.text, contacts, accounts, nil)
			if err != nil {
				if strings.Contains(err.Error(), "API error") || strings.Contains(err.Error(), "rate limit") {
					t.Skipf("ParseTransaction(%q) hit AI: %v", tt.text, err)
				}
				t.Fatalf("ParseTransaction(%q) error = %v", tt.text, err)
			}
			if got.Type != tt.wantTyp {
				t.Errorf("Type = %v, want %v", got.Type, tt.wantTyp)
			}
			if got.SrcID != tt.wantSrc {
				t.Errorf("SrcID = %q, want %q", got.SrcID, tt.wantSrc)
			}
			if got.DstID != tt.wantDst {
				t.Errorf("DstID = %q, want %q", got.DstID, tt.wantDst)
			}
		})
	}
}

// TestParseTransaction_bareWalletMention verifies wallets mentioned without a
// preposition ("salary 50k ebl") are routed by the final transaction type,
// while explicitly keyed wallets keep priority.
func TestParseTransaction_bareWalletMention(t *testing.T) {
	initCache()
	_ = cache.SetCache("salary from office", `{"intent":"income","subcategory_id":"fin-sal"}`, -1)
	_ = cache.SetCache("lunch", `{"intent":"expense","subcategory_id":"food-rest"}`, -1)
	_ = cache.SetCache("dinner", `{"intent":"expense","subcategory_id":"food-rest"}`, -1)
	_ = cache.SetCache("karim", `{"intent":"expense","subcategory_id":"fin-lend"}`, -1)

	contacts := func(name string) bool { return strings.ToLower(name) == "karim" }
	accounts := func(name string) bool {
		switch strings.ToLower(name) {
		case "cash", "ebl", "dbbl":
			return true
		}
		return false
	}

	tests := []struct {
		name        string
		text        string
		wantTyp     models.TransactionType
		wantSrc     string
		wantDst     string
		wantContact string
	}{
		{"income routes to destination", "salary 50k ebl from office", models.IncomeTransaction, "", "ebl", ""},
		{"expense routes to source", "lunch 250 ebl", models.ExpenseTransaction, "ebl", "", ""},
		{"explicit wallet wins", "dinner 400 ebl from dbbl", models.ExpenseTransaction, "dbbl", "", ""},
		{"bare contact resolved", "lent 500 karim", models.ExpenseTransaction, "cash", "", "karim"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransaction(tt.text, contacts, accounts, nil)
			if err != nil {
				if strings.Contains(err.Error(), "API error") || strings.Contains(err.Error(), "rate limit") {
					t.Skipf("ParseTransaction(%q) hit AI: %v", tt.text, err)
				}
				t.Fatalf("ParseTransaction(%q) error = %v", tt.text, err)
			}
			if got.Type != tt.wantTyp {
				t.Errorf("Type = %v, want %v", got.Type, tt.wantTyp)
			}
			if got.SrcID != tt.wantSrc {
				t.Errorf("SrcID = %q, want %q", got.SrcID, tt.wantSrc)
			}
			if got.DstID != tt.wantDst {
				t.Errorf("DstID = %q, want %q", got.DstID, tt.wantDst)
			}
			if got.ContactName != tt.wantContact {
				t.Errorf("ContactName = %q, want %q", got.ContactName, tt.wantContact)
			}
		})
	}
}

// walkBackToDay returns the latest date at or before t with the given day-of-month.
func walkBackToDay(t time.Time, day int) time.Time {
	for t.Day() != day {
		t = t.AddDate(0, 0, -1)
	}
	return t
}

// walkBackToWeekday returns the latest date at or before t falling on wd.
func walkBackToWeekday(t time.Time, wd time.Weekday) time.Time {
	for t.Weekday() != wd {
		t = t.AddDate(0, 0, -1)
	}
	return t
}

// TestParseTransaction_dateTime verifies natural date/time forms — ordinals
// ("on 1st"), weekdays, hour-only times — produce the exact expected
// amount and timestamp. Expected dates are computed independently by
// walking back from today.
func TestParseTransaction_dateTime(t *testing.T) {
	// No cache seeding: every phrase here (internet, lunch, dinner, coffee,
	// rent) must resolve through localClassify without touching the AI.
	initCache()

	loc := pkg.DefaultLocation
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)
	lastFriday := walkBackToWeekday(today, time.Friday)
	first := walkBackToDay(today, 1)
	day31 := walkBackToDay(today, 31)
	jan5 := time.Date(today.Year(), time.January, 5, 0, 0, 0, 0, loc)
	if jan5.After(today) {
		jan5 = jan5.AddDate(-1, 0, 0)
	}
	jan5of2026 := time.Date(2026, time.January, 5, 0, 0, 0, 0, loc)
	at := func(d time.Time, hour, min int) time.Time {
		return time.Date(d.Year(), d.Month(), d.Day(), hour, min, 0, 0, loc)
	}

	tests := []struct {
		name       string
		text       string
		wantAmount float64
		wantTS     time.Time // zero value means "expect roughly now"
	}{
		{"ordinal date", "internet 500 on 1st", 500, first},
		{"ordinal skips short months", "internet 500 on 31st", 500, day31},
		{"bare weekday", "lunch 250 friday", 250, lastFriday},
		{"short weekday after on", "lunch 250 on fri", 250, lastFriday},
		{"today with named time", "lunch 250 today at noon", 250, at(today, 12, 0)},
		{"yesterday defaults to midnight", "dinner 1.5k yesterday", 1500, yesterday},
		{"yesterday with named time", "dinner 400 yesterday at night", 400, at(yesterday, 22, 0)},
		{"bare named time", "dinner 400 night", 400, at(today, 22, 0)},
		{"bare date and named time", "dinner 400 yesterday night", 400, at(yesterday, 22, 0)},
		{"tomorrow", "dinner 500 tomorrow", 500, tomorrow},
		{"hour-only pm", "dinner 400 at 5pm", 400, at(today, 17, 0)},
		{"hour-only pm spaced", "dinner 400 at 5 pm", 400, at(today, 17, 0)},
		{"bare 24h hour", "dinner 400 at 17", 400, at(today, 17, 0)},
		{"clock with minutes", "coffee 100 at 5:30", 100, at(today, 5, 30)},
		{"time before amount", "dinner at 5:30pm 400", 400, at(today, 17, 30)},
		{"date plus clock time", "rent 15k on 1st at 9:15pm", 15000, at(first, 21, 15)},
		{"yearless month day", "internet 500 on jan 5", 500, jan5},
		{"yearless ordinal month day", "internet 500 on jan 5th", 500, jan5},
		{"iso date", "internet 500 on 2026-01-05", 500, jan5of2026},
		{"dd-mm-yyyy date", "internet 500 on 05-01-2026", 500, jan5of2026},
		{"month name with year", "internet 500 on jan 5, 2026", 500, jan5of2026},
		{"bare ordinal is not a date", "internet 500 1st", 500, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransaction(tt.text, nil, nil, nil)
			if err != nil {
				if strings.Contains(err.Error(), "API error") || strings.Contains(err.Error(), "rate limit") {
					t.Skipf("ParseTransaction(%q) hit AI: %v", tt.text, err)
				}
				t.Fatalf("ParseTransaction(%q) error = %v", tt.text, err)
			}
			if got.Amount != tt.wantAmount {
				t.Errorf("Amount = %v, want %v", got.Amount, tt.wantAmount)
			}
			gotTS := time.Unix(got.Timestamp, 0).In(loc)
			if tt.wantTS.IsZero() {
				if d := gotTS.Sub(now); d < -2*time.Minute || d > 2*time.Minute {
					t.Errorf("Timestamp = %v, want about now (%v)", gotTS, now)
				}
				return
			}
			if !gotTS.Equal(tt.wantTS) {
				t.Errorf("Timestamp = %v, want %v", gotTS, tt.wantTS)
			}
		})
	}
}

func TestParseTransaction(t *testing.T) {
	initCache()
	mockContacts := func(name string) bool {
		switch strings.ToLower(name) {
		case "unknown", "masud", "rahim", "ammu", "karim", "rifat":
			return true
		}
		return false
	}

	mockAccounts := func(name string) bool {
		switch strings.ToLower(name) {
		case "brac", "city", "bkash", "dbbl", "ebl", "cash", "nagad":
			return true
		}
		return false
	}
	type args struct {
		texts    []string
		contacts ContactVerifier
		accounts AccountVerifier
	}
	tests := []struct {
		name    string
		args    args
		want    models.Transaction
		wantErr bool
	}{
		{
			name: "scenarios",
			args: args{
				texts: []string{
					"Add 500 taka",
					"Get 500 from masud",
					"cash 500 from ebl",
					"plus 1000",
					"100 taka baksheesh",
					"50 taka rickshaw",
					"ammu ke 5000 pathalam",           // Banglish: "Sent 5000 to Ammu"
					"brac theke city te 10k transfer", // Banglish: "Transfer 10k from BRAC to City"
				},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test",
			args: args{
				texts:    []string{"bought a dress for 500 taka yesterday"},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				texts:    []string{"filed 200k tax"},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				texts:    []string{"bought chocolate for niece 100 taka"},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				texts:    []string{"birthday present for 200"},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				texts:    []string{"gave 500 to MrX note for emergency"},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test5",
			args: args{
				texts:    []string{"Sent 5000 to Ammu on bKash"},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test6",
			args: args{
				texts:    []string{"Transfer 10k from Brac to City"},
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test6",
			args: args{
				texts:    testInputs,
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, text := range tt.args.texts {
				got, err := ParseTransaction(text, tt.args.contacts, tt.args.accounts, nil)
				if (err != nil) != tt.wantErr {
					if strings.Contains(err.Error(), "API error") || strings.Contains(err.Error(), "rate limit") {
						t.Logf("ParseTransaction(%q) failed due to AI API issue: %v", text, err)
					} else {
						t.Errorf("ParseTransaction(%q) error = %v, wantErr %v", text, err, tt.wantErr)
					}
					continue
				}

				fmt.Printf("===========>[ %s ]<===========\n%s\n\n", text, got.Summary(nil))
			}

			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ParseTransaction() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
