package validator

import "testing"

func TestIsValidWalletName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty is allowed", "", true},
		{"plain name", "Personal Savings", true},
		{"internal apostrophe", "Masud's Savings", true},
		{"apostrophe with dash and underscore", "O'Brien-fund_1", true},
		{"leading apostrophe rejected", "'savings", false},
		{"trailing apostrophe rejected", "savings'", false},
		{"leading space rejected", " savings", false},
		{"other special char rejected", "savings!", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsValidWalletName(tc.input); got != tc.want {
				t.Errorf("IsValidWalletName(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestIsValidDisplayName_RejectsApostrophe(t *testing.T) {
	t.Parallel()
	// Contacts use IsValidDisplayName, which must NOT allow an apostrophe.
	if IsValidDisplayName("Masud's") {
		t.Error("IsValidDisplayName should reject apostrophe for contact names")
	}
}
