package transaction

import (
	"fmt"
	"strings"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
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
