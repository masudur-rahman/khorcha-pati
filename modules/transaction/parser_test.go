package transaction

import (
	"strings"
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"

	"github.com/masudur-rahman/go-oneliners"
)

func TestParseTransaction(t *testing.T) {
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
		text     string
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
				text:     "bought a dress for 500 taka yesterday",
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				text:     "filed 200k tax",
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				text:     "bought chocolate for niece 100 taka",
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				text:     "birthday present for 200",
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				text:     "gave 500 to MrX note for emergency",
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				text:     "Sent 5000 to Ammu on bKash",
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				text:     "Transfer 10k from Brac to City",
				contacts: mockContacts,
				accounts: mockAccounts,
			},
			want:    models.Transaction{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransaction(tt.args.text, tt.args.contacts, tt.args.accounts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			oneliners.PrettyJson(got, tt.args.text)
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ParseTransaction() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
