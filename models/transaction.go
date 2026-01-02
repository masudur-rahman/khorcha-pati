package models

type TransactionType string

const (
	ExpenseTransaction  TransactionType = "Expense"
	IncomeTransaction   TransactionType = "Income"
	TransferTransaction TransactionType = "Transfer"

	LoanSubcategoryID   string = "fin-loan"
	LoanRecoverySubID   string = "fin-recover"
	BorrowSubcategoryID string = "fin-borrow"
	BorrowReturnSubID   string = "fin-return"
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
	{ID: "food-rest", Name: "Restaurants", CatID: "food"},
	{ID: "food-groc", Name: "Groceries", CatID: "food"},
	{ID: "food-take", Name: "Takeout", CatID: "food"},
	{ID: "food-snack", Name: "Snacks", CatID: "food"},
	{ID: "food-fruit", Name: "Fruits", CatID: "food"},
	{ID: "food-bev", Name: "Beverages", CatID: "food"},
	{ID: "food-misc", Name: "Cuisine", CatID: "food"},
}

var houseSubs = []TxnSubcategory{
	{ID: "house-rent", Name: "Rent", CatID: "house"},
	{ID: "house-util", Name: "Utilities", CatID: "house"},
	{ID: "house-furn", Name: "Furniture", CatID: "house"},
	{ID: "house-elec", Name: "Electronics", CatID: "house"},
	{ID: "house-real", Name: "Real State", CatID: "house"},
	{ID: "house-misc", Name: "Residence", CatID: "house"},
}

var entSubs = []TxnSubcategory{
	{ID: "ent-movie", Name: "Movies", CatID: "ent"},
	{ID: "ent-sub", Name: "Subscription", CatID: "ent"},
	{ID: "ent-rec", Name: "Recreation", CatID: "ent"},
	{ID: "ent-books", Name: "Books", CatID: "ent"},
	{ID: "ent-misc", Name: "Leisure", CatID: "ent"},
}

var pcSubs = []TxnSubcategory{
	{ID: "pc-salon", Name: "Salon", CatID: "pc"},
	{ID: "pc-toilet", Name: "Toiletries", CatID: "pc"},
	{ID: "pc-gym", Name: "Gym", CatID: "pc"},
	{ID: "pc-cloth", Name: "Clothing", CatID: "pc"},
	{ID: "pc-health", Name: "Health", CatID: "pc"},
	{ID: "pc-med", Name: "Medicine", CatID: "pc"},
	{ID: "pc-access", Name: "Accessories", CatID: "pc"},
	{ID: "pc-misc", Name: "Wellness", CatID: "pc"},
}

var trvSubs = []TxnSubcategory{
	{ID: "trv-accom", Name: "Accommodation", CatID: "trv"},
	{ID: "trv-dine", Name: "Dining", CatID: "trv"},
	{ID: "trv-sight", Name: "Sightseeing", CatID: "trv"},
	{ID: "trv-trans", Name: "Transportation", CatID: "trv"},
	{ID: "trv-gift", Name: "Gifts", CatID: "trv"},
	{ID: "trv-misc", Name: "Journey", CatID: "trv"},
}

var finSubs = []TxnSubcategory{
	{ID: "fin-sal", Name: "Salary", CatID: "fin"},
	{ID: "fin-deposit", Name: "Deposit", CatID: "fin"},
	{ID: "fin-with", Name: "Withdraw", CatID: "fin"},
	{ID: "fin-dps", Name: "DPS", CatID: "fin"},
	{ID: "fin-ccpay", Name: "Credit Card Payment", CatID: "fin"},
	{ID: "fin-bank", Name: "Bank Transfer", CatID: "fin"},
	{ID: "fin-loan", Name: "Loan", CatID: "fin"},
	{ID: "fin-recover", Name: "Loan Recovery", CatID: "fin"},
	{ID: "fin-borrow", Name: "Borrow", CatID: "fin"},
	{ID: "fin-return", Name: "Borrow Return", CatID: "fin"},
	{ID: "fin-tax", Name: "Tax", CatID: "fin"},
	{ID: "fin-charge", Name: "Charges", CatID: "fin"},
	{ID: "fin-flexi", Name: "Mobile Recharge", CatID: "fin"},
	{ID: "fin-misc", Name: "Overhead", CatID: "fin"},
}

var festSubs = []TxnSubcategory{
	{ID: "fest-eid", Name: "Eid", CatID: "fest"},
	{ID: "fest-wed", Name: "Wedding", CatID: "fest"},
	{ID: "fest-others", Name: "Other Festivs", CatID: "fest"},
}

var miscSubs = []TxnSubcategory{
	{ID: "misc-init", Name: "Initial Amount", CatID: "misc"},
	{ID: "misc-give", Name: "Giveaway", CatID: "misc"},
	{ID: "misc-misc", Name: "General", CatID: "misc"},
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
