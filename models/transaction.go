package models

type TransactionType string

const (
	ExpenseTransaction  TransactionType = "Expense"
	IncomeTransaction   TransactionType = "Income"
	TransferTransaction TransactionType = "Transfer"

	// Financial Subcategory IDs for special handling
	LoanReceivedSubID  = "fin-loan"
	LoanRepaymentSubID = "fin-loan-repay"
	LendSubID          = "fin-lend"
	LendRecoverySubID  = "fin-lend-recover"
	BorrowSubID        = "fin-borrow"
	BorrowReturnSubID  = "fin-borrow-return"
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
	{ID: "house", Name: "Housing"},
	{ID: "ent", Name: "Entertainment"},
	{ID: "pc", Name: "Personal Care"},
	{ID: "trv", Name: "Travel"},
	{ID: "fin", Name: "Financial"},
	{ID: "fest", Name: "Festival"},
	{ID: "misc", Name: "Miscellaneous"},
}

var TxnSubcategories []TxnSubcategory

var foodSubs = []TxnSubcategory{
	{ID: "food-restaurant", Name: "Restaurants", CatID: "food"},
	{ID: "food-groceries", Name: "Groceries", CatID: "food"},
	{ID: "food-takeout", Name: "Takeout", CatID: "food"},
	{ID: "food-snacks", Name: "Snacks", CatID: "food"},
	{ID: "food-fruits", Name: "Fruits", CatID: "food"},
	{ID: "food-beverages", Name: "Beverages", CatID: "food"},
	{ID: "food-other", Name: "Other Food", CatID: "food"},
}

var houseSubs = []TxnSubcategory{
	{ID: "house-rent", Name: "Rent", CatID: "house"},
	{ID: "house-utilities", Name: "Utilities", CatID: "house"},
	{ID: "house-furniture", Name: "Furniture", CatID: "house"},
	{ID: "house-electronics", Name: "Electronics", CatID: "house"},
	{ID: "house-realestate", Name: "Real Estate", CatID: "house"},
	{ID: "house-other", Name: "Other Household", CatID: "house"},
}

var entSubs = []TxnSubcategory{
	{ID: "ent-movies", Name: "Movies", CatID: "ent"},
	{ID: "ent-subscription", Name: "Subscription", CatID: "ent"},
	{ID: "ent-recreation", Name: "Recreation", CatID: "ent"},
	{ID: "ent-books", Name: "Books", CatID: "ent"},
	{ID: "ent-other", Name: "Other Entertainment", CatID: "ent"},
}

var pcSubs = []TxnSubcategory{
	{ID: "pc-salon", Name: "Salon", CatID: "pc"},
	{ID: "pc-toiletries", Name: "Toiletries", CatID: "pc"},
	{ID: "pc-gym", Name: "Gym", CatID: "pc"},
	{ID: "pc-clothing", Name: "Clothing", CatID: "pc"},
	{ID: "pc-health", Name: "Health", CatID: "pc"},
	{ID: "pc-medicine", Name: "Medicine", CatID: "pc"},
	{ID: "pc-accessories", Name: "Accessories", CatID: "pc"},
	{ID: "pc-other", Name: "Other Personal Care", CatID: "pc"},
}

var trvSubs = []TxnSubcategory{
	{ID: "trv-accommodation", Name: "Accommodation", CatID: "trv"},
	{ID: "trv-dining", Name: "Dining", CatID: "trv"},
	{ID: "trv-sightseeing", Name: "Sightseeing", CatID: "trv"},
	{ID: "trv-transport", Name: "Transportation", CatID: "trv"},
	{ID: "trv-gifts", Name: "Gifts", CatID: "trv"},
	{ID: "trv-other", Name: "Other Travel", CatID: "trv"},
}

var finSubs = []TxnSubcategory{
	// Income & Regular Transactions
	{ID: "fin-salary", Name: "Salary", CatID: "fin"},
	{ID: "fin-deposit", Name: "Deposit", CatID: "fin"},
	{ID: "fin-withdraw", Name: "Withdrawal", CatID: "fin"},
	{ID: "fin-transfer", Name: "Bank Transfer", CatID: "fin"},
	{ID: "fin-recharge", Name: "Mobile Recharge", CatID: "fin"},

	// Bank Loans
	{ID: "fin-loan", Name: "Loan Received", CatID: "fin"},
	{ID: "fin-loan-repay", Name: "Loan Repayment", CatID: "fin"},

	// Personal Lending (to others)
	{ID: "fin-lend", Name: "Lend", CatID: "fin"},
	{ID: "fin-lend-recover", Name: "Lend Recovery", CatID: "fin"},

	// Personal Borrowing (from others)
	{ID: "fin-borrow", Name: "Borrow", CatID: "fin"},
	{ID: "fin-borrow-return", Name: "Borrow Return", CatID: "fin"},

	// Investments & Charges
	{ID: "fin-dps", Name: "DPS", CatID: "fin"},
	{ID: "fin-cc-payment", Name: "Credit Card Payment", CatID: "fin"},
	{ID: "fin-tax", Name: "Tax", CatID: "fin"},
	{ID: "fin-charges", Name: "Bank Charges", CatID: "fin"},
	{ID: "fin-other", Name: "Other Financial", CatID: "fin"},
}

var festSubs = []TxnSubcategory{
	{ID: "fest-eid", Name: "Eid", CatID: "fest"},
	{ID: "fest-wed", Name: "Wedding", CatID: "fest"},
	{ID: "fest-other", Name: "Other Festivals", CatID: "fest"},
}

var miscSubs = []TxnSubcategory{
	{ID: "misc-initial", Name: "Initial Amount", CatID: "misc"},
	{ID: "misc-giveaway", Name: "Giveaway", CatID: "misc"},
	{ID: "misc-charity", Name: "Charity", CatID: "misc"},
	{ID: "misc-general", Name: "General", CatID: "misc"},
}

func init() {
	TxnSubcategories = append(TxnSubcategories, foodSubs...)
	TxnSubcategories = append(TxnSubcategories, houseSubs...)
	TxnSubcategories = append(TxnSubcategories, entSubs...)
	TxnSubcategories = append(TxnSubcategories, pcSubs...)
	TxnSubcategories = append(TxnSubcategories, trvSubs...)
	TxnSubcategories = append(TxnSubcategories, finSubs...)
	TxnSubcategories = append(TxnSubcategories, festSubs...)
	TxnSubcategories = append(TxnSubcategories, miscSubs...)
}
