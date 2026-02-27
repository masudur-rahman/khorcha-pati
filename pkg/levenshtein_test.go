package pkg_test

import (
	"fmt"
	"strings"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"
)

func ExampleLevenshteinDistance() {
	a := "snack"

	md := 10000
	var rs string

	for _, subcat := range models.TxnSubcategories {
		ld := min(pkg.LevenshteinDistance(subcat.ID, a, true),
			pkg.LevenshteinDistance(subcat.Name, a, true))
		suggestByLevenshtein := ld <= 3
		suggestByPrefix := strings.HasPrefix(strings.ToLower(subcat.Name), strings.ToLower(a))
		if suggestByLevenshtein && ld < md || suggestByPrefix {
			md = ld
			rs = subcat.Name
		}
	}

	fmt.Printf("Distance between `%v` and `%v` is %v\n", a, rs, md)
	// Output: Distance between `snack` and `Snacks` is 1
}
