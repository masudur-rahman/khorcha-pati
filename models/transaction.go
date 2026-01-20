package models

import (
	"fmt"
	"strings"
	"time"
)

var (
	SubCatNameMap      = make(map[string]string)
	SubCatToCatNameMap = make(map[string]string)
)

type TransactionType string

const (
	ExpenseTransaction  TransactionType = "Expense"
	IncomeTransaction   TransactionType = "Income"
	TransferTransaction TransactionType = "Transfer"

	// Financial Subcategory IDs for special handling
	LoanReceivedSubID  = "fin-loan"
	LoanRepaymentSubID = "fin-repay"
	LendSubID          = "fin-lend"
	LendRecoverySubID  = "fin-recover"
	BorrowSubID        = "fin-borrow"
	BorrowReturnSubID  = "fin-return"
)

type Transaction struct {
	ID                 int64 `db:"id,pk autoincr"`
	UserID             int64
	Amount             float64
	SubcategoryID      string
	Type               TransactionType
	SrcID              string
	DstID              string
	DebtorCreditorName string
	Timestamp          int64
	Remarks            string
}

// Summary creates a user-friendly status message
func (t Transaction) Summary() string {
	var sb strings.Builder

	emoji := "💸"
	action := "Expense Recorded"

	switch t.Type {
	case IncomeTransaction:
		emoji = "💰"
		action = "Income Recorded"
	case TransferTransaction:
		emoji = "↔️"
		action = "Transfer Recorded"
	}

	sb.WriteString(fmt.Sprintf("%s *%s*\n", emoji, action))
	sb.WriteString("──────────────\n")

	sb.WriteString(fmt.Sprintf("💵 *Amount:* %.2f\n", t.Amount))

	catName := "Unknown"
	subName := t.SubcategoryID

	if name, exists := SubCatNameMap[t.SubcategoryID]; exists {
		subName = name
	}
	if cat, exists := SubCatToCatNameMap[t.SubcategoryID]; exists {
		catName = cat
	}

	sb.WriteString(fmt.Sprintf("🏷 *Category:* %s › %s\n", catName, subName))

	if t.Type == TransferTransaction {
		sb.WriteString(fmt.Sprintf("🏦 *Flow:* %s ➔ %s\n", formatAccount(t.SrcID), formatAccount(t.DstID)))
	} else if t.Type == IncomeTransaction {
		if t.DstID != "" && t.DstID != "cash" {
			sb.WriteString(fmt.Sprintf("📥 *To:* %s\n", formatAccount(t.DstID)))
		}
		if t.DebtorCreditorName != "" {
			sb.WriteString(fmt.Sprintf("👤 *From:* %s\n", t.DebtorCreditorName))
		}
	} else {
		if t.SrcID != "" && t.SrcID != "cash" {
			sb.WriteString(fmt.Sprintf("💳 *From:* %s\n", formatAccount(t.SrcID)))
		}
		if t.DebtorCreditorName != "" {
			sb.WriteString(fmt.Sprintf("👤 *To:* %s\n", t.DebtorCreditorName))
		}
	}

	ts := time.Unix(t.Timestamp, 0)
	sb.WriteString(fmt.Sprintf("📅 *Date:* %s\n", ts.Format("02 Jan, 2006 • 03:04 PM")))

	if t.Remarks != "" && t.Remarks != t.SubcategoryID {
		sb.WriteString(fmt.Sprintf("📝 *Note:* %s\n", t.Remarks))
	}

	return sb.String()
}

func formatAccount(id string) string {
	if id == "" {
		return "Unknown"
	}
	return strings.ToUpper(id)
}

type TxnCategory struct {
	ID   string `db:",pk"`
	Name string
}

type TxnSubcategory struct {
	ID    string `db:",pk"`
	Name  string
	CatID string
}

var TxnCategories = []TxnCategory{
	{ID: "food", Name: "Food"},
	{ID: "trans", Name: "Transport"},
	{ID: "shop", Name: "Shopping"},
	{ID: "fin", Name: "Financial"},
	{ID: "house", Name: "Housing"},
	{ID: "health", Name: "Health"},
	{ID: "pc", Name: "Personal Care"},
	{ID: "fam", Name: "Family"},
	{ID: "edu", Name: "Education"},
	{ID: "ent", Name: "Entertainment"},
	{ID: "trv", Name: "Travel"},
	{ID: "fest", Name: "Festival"},
	{ID: "misc", Name: "Miscellaneous"},
}

var TxnSubcategories []TxnSubcategory

var foodSubs = []TxnSubcategory{
	{ID: "food-groc", Name: "Grocery", CatID: "food"},
	{ID: "food-veg", Name: "Vegetables", CatID: "food"},
	{ID: "food-fruit", Name: "Fruits", CatID: "food"},
	{ID: "food-fish", Name: "Fish", CatID: "food"},
	{ID: "food-meat", Name: "Meat", CatID: "food"},
	{ID: "food-dairy", Name: "Dairy & Eggs", CatID: "food"},
	{ID: "food-bakery", Name: "Bakery", CatID: "food"},
	{ID: "food-rest", Name: "Restaurant", CatID: "food"},
	{ID: "food-street", Name: "Street Food", CatID: "food"},
	{ID: "food-take", Name: "Takeout", CatID: "food"},
	{ID: "food-snack", Name: "Snacks", CatID: "food"},
	{ID: "food-bev", Name: "Beverages", CatID: "food"},
	{ID: "food-misc", Name: "General Food", CatID: "food"},
}

var transSubs = []TxnSubcategory{
	{ID: "trans-pub", Name: "Bus/Train", CatID: "trans"},
	{ID: "trans-taxi", Name: "Taxi/Ride", CatID: "trans"},
	{ID: "trans-fuel", Name: "Fuel", CatID: "trans"},
	{ID: "trans-toll", Name: "Tolls/Parking", CatID: "trans"},
	{ID: "trans-maint", Name: "Vehicle Maint", CatID: "trans"},
	{ID: "trans-other", Name: "Other Transport", CatID: "trans"},
}

var shopSubs = []TxnSubcategory{
	{ID: "shop-supply", Name: "Household", CatID: "shop"},
	{ID: "shop-cloth", Name: "Clothing", CatID: "shop"},
	{ID: "shop-foot", Name: "Footwear", CatID: "shop"},
	{ID: "shop-elec", Name: "Electronics", CatID: "shop"},
	{ID: "shop-jewelry", Name: "Jewelry", CatID: "shop"},
	{ID: "shop-beauty", Name: "Cosmetics", CatID: "shop"},
	{ID: "shop-acc", Name: "Accessories", CatID: "shop"},
	{ID: "shop-stat", Name: "Stationary", CatID: "shop"},
	{ID: "shop-other", Name: "General Shopping", CatID: "shop"},
}

var finSubs = []TxnSubcategory{
	{ID: "fin-sal", Name: "Salary", CatID: "fin"},
	{ID: "fin-prof", Name: "Profit/Bonus", CatID: "fin"},
	{ID: "fin-interest", Name: "Interest", CatID: "fin"},
	{ID: "fin-deposit", Name: "Deposit", CatID: "fin"},
	{ID: "fin-with", Name: "Withdraw", CatID: "fin"},
	{ID: "fin-transfer", Name: "Acc Transfer", CatID: "fin"},
	{ID: "fin-flexi", Name: "Mobile Recharge", CatID: "fin"},
	{ID: "fin-ccpay", Name: "Credit Card Payment", CatID: "fin"},
	{ID: "fin-dps", Name: "DPS", CatID: "fin"},
	{ID: "fin-loan", Name: "Bank Loan", CatID: "fin"},
	{ID: "fin-repay", Name: "Bank Repayment", CatID: "fin"},
	{ID: "fin-lend", Name: "Lending", CatID: "fin"},
	{ID: "fin-recover", Name: "Lend Recovery", CatID: "fin"},
	{ID: "fin-borrow", Name: "Borrowing", CatID: "fin"},
	{ID: "fin-return", Name: "Borrow Return", CatID: "fin"},
	{ID: "fin-tax", Name: "VAT/Tax", CatID: "fin"},
	{ID: "fin-charge", Name: "Charges", CatID: "fin"},
	{ID: "fin-ins", Name: "Insurance", CatID: "fin"},
	{ID: "fin-gold", Name: "Gold Investment", CatID: "fin"},
	{ID: "fin-invest", Name: "Stocks/Assets", CatID: "fin"},
	{ID: "fin-misc", Name: "Overhead", CatID: "fin"},
}

var houseSubs = []TxnSubcategory{
	{ID: "house-rent", Name: "Rent", CatID: "house"},
	{ID: "house-util", Name: "Utilities", CatID: "house"},
	{ID: "house-net", Name: "Internet", CatID: "house"},
	{ID: "house-serv", Name: "Maid/Service", CatID: "house"},
	{ID: "house-maint", Name: "Maintenance", CatID: "house"},
	{ID: "house-furn", Name: "Furniture", CatID: "house"},
	{ID: "house-real", Name: "Real Estate", CatID: "house"},
	{ID: "house-misc", Name: "General Household", CatID: "house"},
}

var healthSubs = []TxnSubcategory{
	{ID: "health-doc", Name: "Doctor Visit", CatID: "health"},
	{ID: "health-test", Name: "Medical Tests", CatID: "health"},
	{ID: "health-med", Name: "Medicine", CatID: "health"},
	{ID: "health-other", Name: "Other Health Exp", CatID: "health"},
}

var pcSubs = []TxnSubcategory{
	{ID: "pc-salon", Name: "Salon", CatID: "pc"},
	{ID: "pc-skin", Name: "Skincare", CatID: "pc"},
	{ID: "pc-spa", Name: "Spa & Massage", CatID: "pc"},
	{ID: "pc-toilet", Name: "Toiletries", CatID: "pc"},
	{ID: "pc-fit", Name: "Fitness", CatID: "pc"},
	{ID: "pc-misc", Name: "Wellness", CatID: "pc"},
}

var famSubs = []TxnSubcategory{
	{ID: "fam-allow", Name: "Spouse Allowance", CatID: "fam"},
	{ID: "fam-par", Name: "Parents", CatID: "fam"},
	{ID: "fam-baby", Name: "Baby Needs", CatID: "fam"},
	{ID: "fam-child", Name: "Kids Needs", CatID: "fam"},
	{ID: "fam-care", Name: "Family Care", CatID: "fam"},
	{ID: "fam-other", Name: "Other Family Exp", CatID: "fam"},
}

var eduSubs = []TxnSubcategory{
	{ID: "edu-course", Name: "Courses", CatID: "edu"},
	{ID: "edu-book", Name: "Books/Stationary", CatID: "edu"},
	{ID: "edu-exam", Name: "Exam Fees", CatID: "edu"},
	{ID: "edu-other", Name: "Other Education", CatID: "edu"},
}

var entSubs = []TxnSubcategory{
	{ID: "ent-movie", Name: "Movies", CatID: "ent"},
	{ID: "ent-sub", Name: "Subscription", CatID: "ent"},
	{ID: "ent-rec", Name: "Recreation", CatID: "ent"},
	{ID: "ent-game", Name: "Gaming", CatID: "ent"},
	{ID: "ent-event", Name: "Concerts/Events", CatID: "ent"},
	{ID: "ent-misc", Name: "Hobby/Misc", CatID: "ent"},
}

var trvSubs = []TxnSubcategory{
	{ID: "trv-ticket", Name: "Tickets", CatID: "trv"},
	{ID: "trv-hotel", Name: "Hotel", CatID: "trv"},
	{ID: "trv-dine", Name: "Dining", CatID: "trv"},
	{ID: "trv-sight", Name: "Sightseeing", CatID: "trv"},
	{ID: "trv-trans", Name: "Transportation", CatID: "trv"},
	{ID: "trv-gift", Name: "Gifts", CatID: "trv"},
	{ID: "trv-misc", Name: "Journey", CatID: "trv"},
}

var festSubs = []TxnSubcategory{
	{ID: "fest-eid", Name: "Eid", CatID: "fest"},
	{ID: "fest-wed", Name: "Wedding", CatID: "fest"},
	{ID: "fest-others", Name: "Other Festivs", CatID: "fest"},
	{ID: "fest-gift", Name: "Gifts", CatID: "fest"},
	{ID: "fest-decor", Name: "Decoration", CatID: "fest"},
	{ID: "fest-charity", Name: "Zakat/Donation", CatID: "fest"},
	{ID: "fest-food", Name: "Fest Feast", CatID: "fest"},
}

var miscSubs = []TxnSubcategory{
	{ID: "misc-init", Name: "Initial Amount", CatID: "misc"},
	{ID: "misc-gift", Name: "General Gifts", CatID: "misc"},
	{ID: "misc-charity", Name: "General Charity", CatID: "misc"},
	{ID: "misc-office", Name: "Office/Work Exp", CatID: "misc"},
	{ID: "misc-loss", Name: "Lost/Stolen", CatID: "misc"},
	{ID: "misc-adj", Name: "Balance Adjustment", CatID: "misc"},
	{ID: "misc-misc", Name: "General", CatID: "misc"},
}

func init() {
	TxnSubcategories = append(TxnSubcategories, foodSubs...)
	TxnSubcategories = append(TxnSubcategories, transSubs...)
	TxnSubcategories = append(TxnSubcategories, shopSubs...)
	TxnSubcategories = append(TxnSubcategories, finSubs...)
	TxnSubcategories = append(TxnSubcategories, houseSubs...)
	TxnSubcategories = append(TxnSubcategories, healthSubs...)
	TxnSubcategories = append(TxnSubcategories, pcSubs...)
	TxnSubcategories = append(TxnSubcategories, famSubs...)
	TxnSubcategories = append(TxnSubcategories, eduSubs...)
	TxnSubcategories = append(TxnSubcategories, entSubs...)
	TxnSubcategories = append(TxnSubcategories, trvSubs...)
	TxnSubcategories = append(TxnSubcategories, festSubs...)
	TxnSubcategories = append(TxnSubcategories, miscSubs...)

	catIDToName := make(map[string]string)
	for _, cat := range TxnCategories {
		catIDToName[cat.ID] = cat.Name
	}

	for _, sub := range TxnSubcategories {
		SubCatNameMap[sub.ID] = sub.Name

		if catName, ok := catIDToName[sub.CatID]; ok {
			SubCatToCatNameMap[sub.ID] = catName
		} else {
			SubCatToCatNameMap[sub.ID] = "Unknown"
		}
	}
}
