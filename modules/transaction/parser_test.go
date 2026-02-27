package transaction

import (
	"fmt"
	"strings"
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"
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
	"paid 1500 for wifi bill january",
	"dinner at kfc with friends 1200",
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
		Type: cache.CacheRedis,
		Redis: cache.ConfigRedis{
			Host: "localhost",
			Port: "6379",
		},
	}
	cache.Init(cfg)
}

func TestParseTransaction(t *testing.T) {
	initCache()
	mockContacts := func(name string) bool {
		return strings.ToLower(name) == "unknown"
	}

	mockAccounts := func(name string) bool {
		switch strings.ToLower(name) {
		case "brac", "city", "bkash", "dbbl":
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
				got, err := ParseTransaction(text, tt.args.contacts, tt.args.accounts)
				if (err != nil) != tt.wantErr {
					t.Errorf("ParseTransaction() error = %v, wantErr %v", err, tt.wantErr)
					//return
					continue
				}

				fmt.Printf("===========>[ %s ]<===========\n%s\n\n", text, got.Summary())
				//oneliners.PrettyJson(got.Summary(), text)
			}

			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ParseTransaction() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
