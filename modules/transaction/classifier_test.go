package transaction

import (
	"errors"
	"testing"
)

func TestNormalizePhrase(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"  Lunch  ", "lunch"},
		{"had a Lunch", "lunch"},
		{"Salary!!", "salary"},
		{"gave back the money", "gave back money"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := normalizePhrase(tt.in); got != tt.want {
			t.Errorf("normalizePhrase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLocalClassify(t *testing.T) {
	tests := []struct {
		name   string
		phrase string
		want   string
		wantOK bool
	}{
		{"salary", "got my salary", "fin-sal", true},
		{"withdraw", "atm withdraw", "fin-with", true},
		{"multiword beats single", "credit card bill", "fin-ccpay", true},
		{"dinner", "dinner", "food-rest", true},
		{"lunch", "lunch", "food-rest", true},
		{"breakfast", "breakfast", "food-rest", true},
		{"groceries", "groceries", "food-groc", true},
		{"taxi", "taxi", "trans-taxi", true},
		{"hospital", "hospital", "health-doc", true},
		{"bare water is beverage", "water", "food-bev", true},
		{"water bill is utility", "water bill", "house-util", true},
		{"bare mobile is electronics", "mobile", "shop-elec", true},
		{"mobile recharge is flexi", "mobile recharge", "fin-flexi", true},
		{"no match", "asdf qwer zxcv", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := localClassify(tt.phrase)
			if ok != tt.wantOK {
				t.Fatalf("localClassify(%q) ok = %v, want %v (got %q)", tt.phrase, ok, tt.wantOK, got)
			}
			if ok && got != tt.want {
				t.Errorf("localClassify(%q) = %q, want %q", tt.phrase, got, tt.want)
			}
		})
	}
}

func TestIsRateLimitErr(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("googleapi: Error 429: RESOURCE_EXHAUSTED"), true},
		{errors.New("rate limit exceeded"), true},
		{errors.New("too many requests"), true},
		{errors.New("invalid api key"), false},
	}
	for _, tt := range tests {
		if got := isRateLimitErr(tt.err); got != tt.want {
			t.Errorf("isRateLimitErr(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}
