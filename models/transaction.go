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

	// Separator is the standard visual divider for Telegram messages.
	Separator = "──────────────"
)

// TxnListQuery scopes and paginates a transaction listing. Zero-value fields are
// ignored. Page is 1-based; Limit <= 0 returns all matches (no pagination).
type TxnListQuery struct {
	UserID  int64
	Type    TransactionType
	Start   int64  // inclusive timestamp lower bound (0 = unbounded)
	End     int64  // inclusive timestamp upper bound (0 = unbounded)
	Wallet  string // matches src_id OR dst_id
	Contact string // matches contact_name
	Page    int64
	Limit   int64
}

type Transaction struct {
	ID            int64           `db:"id,pk autoincr" json:"id"`
	UserID        int64           `json:"userId"`
	Amount        float64         `json:"amount"`
	SubcategoryID string          `json:"subcategoryId"`
	Type          TransactionType `json:"type"`
	SrcID         string          `json:"srcId"`
	DstID         string          `json:"dstId"`
	ContactName   string          `json:"contactName"`
	Timestamp     int64           `json:"timestamp"`
	Remarks       string          `json:"remarks"`

	DeletedAt int64 `db:"deleted_at,req" json:"deletedAt"` // 0 = active; non-zero = unix timestamp of soft-delete
	CreatedAt int64 `db:"created_at" json:"createdAt"`     // unix timestamp of creation
}

func (Transaction) TableName() string {
	return "transaction"
}

// Summary creates a user-friendly status message
func (t Transaction) Summary(loc *time.Location) string {
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
	sb.WriteString(Separator + "\n")

	sb.WriteString(fmt.Sprintf("💵 *Amount:* %s\n", FormatMoneySigned(t.Amount, t.Type)))

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
		if t.ContactName != "" {
			sb.WriteString(fmt.Sprintf("👤 *From:* %s\n", t.ContactName))
		}
	} else {
		if t.SrcID != "" && t.SrcID != "cash" {
			sb.WriteString(fmt.Sprintf("💳 *From:* %s\n", formatAccount(t.SrcID)))
		}
		if t.ContactName != "" {
			sb.WriteString(fmt.Sprintf("👤 *To:* %s\n", t.ContactName))
		}
	}

	if loc == nil {
		loc = time.Local
	}
	ts := time.Unix(t.Timestamp, 0).In(loc)
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
	ID   string `db:",pk" json:"id"`
	Name string `json:"name"`
}

func (TxnCategory) TableName() string {
	return "txn_category"
}

type TxnSubcategory struct {
	ID    string `db:",pk" json:"id"`
	Name  string `json:"name"`
	CatID string `json:"catId"`
	// Hint (rich AI context) and Keywords (short UI search terms) get real columns:
	// styx has no skip tag (db:"-" collides on a shared "_" column). Values served
	// to clients are always sourced from the in-memory taxonomy via handler enrichment.
	Hint     string `json:"hint,omitempty"`
	Keywords string `json:"keywords,omitempty"`
}

// ContainsType reports whether typ is present in the given slice.
func ContainsType(types []TransactionType, typ TransactionType) bool {
	for _, t := range types {
		if t == typ {
			return true
		}
	}
	return false
}

var (
	typExpense  = []TransactionType{ExpenseTransaction}
	typIncome   = []TransactionType{IncomeTransaction}
	typTransfer = []TransactionType{TransferTransaction}
	typExpOrInc = []TransactionType{ExpenseTransaction, IncomeTransaction}
)

// SubcategoryByID indexes every subcategory for O(1) lookup. Populated in init().
var SubcategoryByID = make(map[string]TxnSubcategory)

// SubcategoryTypes maps a subcategory ID to the transaction types it belongs to.
var SubcategoryTypes = make(map[string][]TransactionType)

// CategoryTypes maps a category ID to the union of types across its subcategories.
var CategoryTypes = make(map[string][]TransactionType)

func (TxnSubcategory) TableName() string {
	return "txn_subcategory"
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
	{ID: "food-groc", Name: "Grocery", CatID: "food", Hint: "rice, flour, oil, spices, lentils, sugar, salt, pantry staples"},
	{ID: "food-veg", Name: "Vegetables", CatID: "food", Hint: "potatoes, onions, tomatoes, greens, fresh vegetables"},
	{ID: "food-fruit", Name: "Fruits", CatID: "food", Hint: "banana, mango, apple, seasonal fruits"},
	{ID: "food-fish", Name: "Fish", CatID: "food", Hint: "ilish, rui, tilapia, shrimp, dried fish"},
	{ID: "food-meat", Name: "Meat", CatID: "food", Hint: "chicken, beef, mutton, kurbani meat"},
	{ID: "food-dairy", Name: "Dairy & Eggs", CatID: "food", Hint: "milk, eggs, yogurt, cheese, butter, ghee"},
	{ID: "food-bakery", Name: "Bakery", CatID: "food", Hint: "bread, cake, biscuits, pastry, pitha"},
	{ID: "food-rest", Name: "Restaurant", CatID: "food", Hint: "dine-in meals, biryani, set menu, buffet"},
	{ID: "food-street", Name: "Street Food", CatID: "food", Hint: "fuchka, chotpoti, jhalmuri, singara, samosa, shingara"},
	{ID: "food-take", Name: "Takeout", CatID: "food", Hint: "food delivery, foodpanda, pathao food, takeaway"},
	{ID: "food-snack", Name: "Snacks", CatID: "food", Hint: "chips, chanachur, nuts, nimki, packaged snacks"},
	{ID: "food-bev", Name: "Beverages", CatID: "food", Hint: "tea, coffee, juice, water, soft drinks, lassi, sherbet"},
	{ID: "food-misc", Name: "General Food", CatID: "food", Hint: "any food item not covered by other food subcategories"},
}

var transSubs = []TxnSubcategory{
	{ID: "trans-pub", Name: "Bus/Train", CatID: "trans", Hint: "bus fare, train ticket, metro, local transport"},
	{ID: "trans-taxi", Name: "Taxi/Ride", CatID: "trans", Hint: "uber, pathao, rickshaw, cng, auto-rickshaw, bike ride"},
	{ID: "trans-fuel", Name: "Fuel", CatID: "trans", Hint: "petrol, diesel, octane, gas station"},
	{ID: "trans-toll", Name: "Tolls/Parking", CatID: "trans", Hint: "bridge toll, highway toll, parking fee"},
	{ID: "trans-maint", Name: "Vehicle Maint", CatID: "trans", Hint: "bike/car servicing, tire, oil change, repair, mechanic"},
	{ID: "trans-other", Name: "Other Transport", CatID: "trans", Hint: "any transport cost not covered above"},
}

var shopSubs = []TxnSubcategory{
	{ID: "shop-supply", Name: "Household", CatID: "shop", Hint: "cleaning supplies, detergent, dish soap, broom, bucket"},
	{ID: "shop-cloth", Name: "Clothing", CatID: "shop", Hint: "shirt, pant, lungi, saree, t-shirt, genji, sando, underwear, socks"},
	{ID: "shop-foot", Name: "Footwear", CatID: "shop", Hint: "shoes, sandals, slippers, boots"},
	{ID: "shop-elec", Name: "Electronics", CatID: "shop", Hint: "phone, charger, earphone, laptop, cable, adapter, gadget"},
	{ID: "shop-jewelry", Name: "Jewelry", CatID: "shop", Hint: "gold chain, ring, earring, necklace, bangle"},
	{ID: "shop-beauty", Name: "Cosmetics", CatID: "shop", Hint: "makeup, lipstick, foundation, perfume, lotion"},
	{ID: "shop-acc", Name: "Accessories", CatID: "shop", Hint: "watch, bag, wallet, belt, sunglasses, umbrella"},
	{ID: "shop-stat", Name: "Stationary", CatID: "shop", Hint: "pen, notebook, paper, file, office supplies"},
	{ID: "shop-other", Name: "General Shopping", CatID: "shop", Hint: "any purchase not covered by other shopping subcategories"},
}

var finSubs = []TxnSubcategory{
	{ID: "fin-sal", Name: "Salary", CatID: "fin", Hint: "monthly salary, wages, paycheck"},
	{ID: "fin-prof", Name: "Profit/Bonus", CatID: "fin", Hint: "bonus, profit share, freelance income, cashback"},
	{ID: "fin-interest", Name: "Interest", CatID: "fin", Hint: "bank interest, savings interest, FDR interest"},
	{ID: "fin-deposit", Name: "Deposit", CatID: "fin", Hint: "bank deposit, FDR, savings deposit"},
	{ID: "fin-with", Name: "Withdraw", CatID: "fin", Hint: "ATM withdrawal, bank withdrawal, cashout, cash out, cashed out, withdrew cash"},
	{ID: "fin-transfer", Name: "Acc Transfer", CatID: "fin", Hint: "bkash, nagad, rocket, bank-to-bank, send money"},
	{ID: "fin-flexi", Name: "Mobile Recharge", CatID: "fin", Hint: "flexiload, mobile top-up, airtime, data pack"},
	{ID: "fin-ccpay", Name: "Credit Card Payment", CatID: "fin", Hint: "credit card bill, CC payment"},
	{ID: "fin-dps", Name: "DPS", CatID: "fin", Hint: "deposit pension scheme, monthly savings plan"},
	{ID: "fin-loan", Name: "Bank Loan", CatID: "fin", Hint: "bank loan received, took a loan from bank, institutional loan, loan disbursed, got a loan"},
	{ID: "fin-repay", Name: "Bank Repayment", CatID: "fin", Hint: "loan EMI, bank installment payment, monthly installment, paid loan installment, mortgage payment, repaid bank loan"},
	{ID: "fin-lend", Name: "Lending", CatID: "fin", Hint: "lend money, lent to friend, gave a loan, loaned money to someone, gave money to, advance to someone, dhar dilam"},
	{ID: "fin-recover", Name: "Lend Recovery", CatID: "fin", Hint: "recovered money i lent, got my money back, they paid me back, collected the debt, loan returned to me, dhar ferot pelam"},
	{ID: "fin-borrow", Name: "Borrowing", CatID: "fin", Hint: "borrowed money, took a loan from a friend, got money from someone, personal borrowing, dhar nilam, ধার নিলাম"},
	{ID: "fin-return", Name: "Borrow Return", CatID: "fin", Hint: "returned borrowed money, paid back, gave back the money, settled my debt, repaid a friend, returned the loan, paying back what i owe, dhar shodh, ধার শোধ"},
	{ID: "fin-tax", Name: "VAT/Tax", CatID: "fin", Hint: "income tax, VAT, govt fees, stamp duty"},
	{ID: "fin-charge", Name: "Charges", CatID: "fin", Hint: "bank charges, service fee, transaction fee, penalty"},
	{ID: "fin-ins", Name: "Insurance", CatID: "fin", Hint: "life insurance, health insurance, vehicle insurance premium"},
	{ID: "fin-gold", Name: "Gold Investment", CatID: "fin", Hint: "gold purchase, gold bar, gold coin"},
	{ID: "fin-invest", Name: "Stocks/Assets", CatID: "fin", Hint: "share market, mutual fund, crypto, bond, investment"},
	{ID: "fin-misc", Name: "Overhead", CatID: "fin", Hint: "any financial transaction not covered above"},
}

var houseSubs = []TxnSubcategory{
	{ID: "house-rent", Name: "Rent", CatID: "house", Hint: "monthly house rent, basha bhara"},
	{ID: "house-util", Name: "Utilities", CatID: "house", Hint: "electricity, gas, water bill, DESCO, DPDC"},
	{ID: "house-net", Name: "Internet", CatID: "house", Hint: "WiFi, broadband, ISP bill"},
	{ID: "house-serv", Name: "Maid/Service", CatID: "house", Hint: "maid salary, cleaner, kajer bua, driver salary"},
	{ID: "house-maint", Name: "Maintenance", CatID: "house", Hint: "plumbing, electrician, house repair, paint, service charge"},
	{ID: "house-furn", Name: "Furniture", CatID: "house", Hint: "table, chair, bed, shelf, curtain, home decor"},
	{ID: "house-real", Name: "Real Estate", CatID: "house", Hint: "flat purchase, land, plot, construction, registry"},
	{ID: "house-misc", Name: "General Household", CatID: "house", Hint: "any housing cost not covered above"},
}

var healthSubs = []TxnSubcategory{
	{ID: "health-doc", Name: "Doctor Visit", CatID: "health", Hint: "doctor fee, consultation, clinic visit, hospital"},
	{ID: "health-test", Name: "Medical Tests", CatID: "health", Hint: "blood test, X-ray, ultrasound, lab test, diagnostic"},
	{ID: "health-med", Name: "Medicine", CatID: "health", Hint: "pharmacy, tablets, syrup, ointment, prescription drugs"},
	{ID: "health-other", Name: "Other Health Exp", CatID: "health", Hint: "surgery, therapy, dental, eye care, ambulance"},
}

var pcSubs = []TxnSubcategory{
	{ID: "pc-salon", Name: "Salon", CatID: "pc", Hint: "haircut, shave, barber, parlor, grooming"},
	{ID: "pc-skin", Name: "Skincare", CatID: "pc", Hint: "face wash, sunscreen, cream, moisturizer"},
	{ID: "pc-spa", Name: "Spa & Massage", CatID: "pc", Hint: "massage, spa treatment, body care"},
	{ID: "pc-toilet", Name: "Toiletries", CatID: "pc", Hint: "soap, shampoo, toothpaste, razor, tissue, sanitary"},
	{ID: "pc-fit", Name: "Fitness", CatID: "pc", Hint: "gym, yoga, workout, swimming, sports club"},
	{ID: "pc-smoke", Name: "Smoking/Habits", CatID: "pc", Hint: "cigarettes, vape, betel leaf, paan, supari, gutka"},
	{ID: "pc-misc", Name: "Wellness", CatID: "pc", Hint: "any personal care not covered above"},
}

var famSubs = []TxnSubcategory{
	{ID: "fam-allow", Name: "Spouse Allowance", CatID: "fam", Hint: "money given to wife/husband, pocket money for spouse"},
	{ID: "fam-par", Name: "Parents", CatID: "fam", Hint: "money sent to parents, baba-ma, family support"},
	{ID: "fam-baby", Name: "Baby Needs", CatID: "fam", Hint: "diaper, baby food, formula, baby clothes"},
	{ID: "fam-child", Name: "Kids Needs", CatID: "fam", Hint: "school fees, tuition, toys, kids clothing, pocket money"},
	{ID: "fam-care", Name: "Family Care", CatID: "fam", Hint: "elder care, relatives support, family medical"},
	{ID: "fam-other", Name: "Other Family Exp", CatID: "fam", Hint: "any family expense not covered above"},
}

var eduSubs = []TxnSubcategory{
	{ID: "edu-course", Name: "Courses", CatID: "edu", Hint: "online course, udemy, coaching, training, tuition class"},
	{ID: "edu-book", Name: "Books/Stationary", CatID: "edu", Hint: "textbooks, reference books, study materials"},
	{ID: "edu-exam", Name: "Exam Fees", CatID: "edu", Hint: "exam registration, admission test, certificate fee"},
	{ID: "edu-other", Name: "Other Education", CatID: "edu", Hint: "any education expense not covered above"},
}

var entSubs = []TxnSubcategory{
	{ID: "ent-movie", Name: "Movies", CatID: "ent", Hint: "cinema, movie ticket, Netflix watch party"},
	{ID: "ent-sub", Name: "Subscription", CatID: "ent", Hint: "Netflix, YouTube Premium, Spotify, app subscription"},
	{ID: "ent-rec", Name: "Recreation", CatID: "ent", Hint: "park, zoo, amusement, outing, picnic, hangout, date"},
	{ID: "ent-game", Name: "Gaming", CatID: "ent", Hint: "video games, mobile games, in-app purchase, PS/Xbox"},
	{ID: "ent-event", Name: "Concerts/Events", CatID: "ent", Hint: "concert ticket, sports event, cultural event"},
	{ID: "ent-misc", Name: "Hobby/Misc", CatID: "ent", Hint: "hobbies, photography, music, art supplies"},
}

var trvSubs = []TxnSubcategory{
	{ID: "trv-ticket", Name: "Tickets", CatID: "trv", Hint: "flight, bus, train, launch ticket for travel/trip"},
	{ID: "trv-hotel", Name: "Hotel", CatID: "trv", Hint: "hotel, resort, airbnb, accommodation during trip"},
	{ID: "trv-dine", Name: "Dining", CatID: "trv", Hint: "meals during travel, restaurant on trip"},
	{ID: "trv-sight", Name: "Sightseeing", CatID: "trv", Hint: "entry fees, tour guide, attraction tickets"},
	{ID: "trv-trans", Name: "Transportation", CatID: "trv", Hint: "local transport during travel, rental car, boat"},
	{ID: "trv-gift", Name: "Gifts", CatID: "trv", Hint: "souvenirs, gifts bought during travel"},
	{ID: "trv-misc", Name: "Journey", CatID: "trv", Hint: "any travel expense not covered above"},
}

var festSubs = []TxnSubcategory{
	{ID: "fest-eid", Name: "Eid", CatID: "fest", Hint: "eid shopping, eidi, eid preparation, salami"},
	{ID: "fest-wed", Name: "Wedding", CatID: "fest", Hint: "wedding gift, biye, walima, wedding expenses"},
	{ID: "fest-others", Name: "Other Festivs", CatID: "fest", Hint: "puja, christmas, new year, pohela boishakh, milad"},
	{ID: "fest-gift", Name: "Gifts", CatID: "fest", Hint: "festival gifts, birthday gift, celebration gift"},
	{ID: "fest-decor", Name: "Decoration", CatID: "fest", Hint: "lights, flowers, balloons, banners, festive decor"},
	{ID: "fest-charity", Name: "Zakat/Donation", CatID: "fest", Hint: "zakat, fitra, sadaqah, donation, daan"},
	{ID: "fest-food", Name: "Fest Feast", CatID: "fest", Hint: "feast, special meals, dawat, iftar party, sehri"},
}

var miscSubs = []TxnSubcategory{
	{ID: "misc-init", Name: "Initial Amount", CatID: "misc", Hint: "opening balance, starting amount when adding wallet"},
	{ID: "misc-gift", Name: "General Gifts", CatID: "misc", Hint: "non-festival gift, birthday present, surprise gift"},
	{ID: "misc-charity", Name: "General Charity", CatID: "misc", Hint: "non-festival charity, help, tip, baksheesh"},
	{ID: "misc-office", Name: "Office/Work Exp", CatID: "misc", Hint: "office lunch, printing, courier, work-related expense"},
	{ID: "misc-loss", Name: "Lost/Stolen", CatID: "misc", Hint: "lost money, theft, pickpocket, damaged item"},
	{ID: "misc-adj", Name: "Balance Adjustment", CatID: "misc", Hint: "correction entry, balance fix, rounding adjustment"},
	{ID: "misc-misc", Name: "General", CatID: "misc", Hint: "LAST RESORT only — use when no other subcategory fits at all"},
}

// subKeywords maps a subcategory ID to short, realistic search terms a user would
// actually type in the UI. Kept separate from Hint (which stays verbose for the AI).
var subKeywords = map[string]string{
	// Food
	"food-groc": "grocery, groceries, bazar, bajar, monihari, rice, oil, sugar, salt, flour", "food-veg": "veggies, shobji, torkari, potato, onion",
	"food-fruit": "fruit, fol, banana, mango, apple", "food-fish": "fish, mach, maach, ilish",
	"food-meat": "meat, mangsho, gosht, chicken, beef, mutton", "food-dairy": "milk, egg, eggs, yogurt, cheese, butter, ghee",
	"food-bakery": "bread, cake, biscuit, pastry, pitha", "food-rest": "restaurant, dining, biryani, kacchi, buffet, lunch, dinner, breakfast",
	"food-street": "street food, fuchka, chotpoti, jhalmuri, singara, shingara, samosa", "food-take": "takeout, takeaway, foodpanda, delivery",
	"food-snack": "snacks, snack, nasta, tiffin, chips, chanachur", "food-bev": "tea, cha, coffee, juice, water, drinks",
	"food-misc": "food, khabar",
	// Transport
	"trans-pub": "bus, train, metro, launch, leguna, tempo", "trans-taxi": "taxi, uber, pathao, rickshaw, cng",
	"trans-fuel": "fuel, petrol, diesel, octane, gas", "trans-toll": "toll, parking",
	"trans-maint": "servicing, repair, mechanic", "trans-other": "transport, fare",
	// Shopping
	"shop-supply": "detergent, cleaning, broom", "shop-cloth": "clothes, jama, kapor, panjabi, lungi, shirt, pant, saree",
	"shop-foot": "shoes, shoe, juta, sandal, slipper", "shop-elec": "phone, mobile, charger, earphone, headphone, laptop, gadget",
	"shop-jewelry": "jewelry, gold, ring, chain", "shop-beauty": "makeup, lipstick, lotion, perfume, cosmetics",
	"shop-acc": "watch, bag, belt, sunglasses, umbrella", "shop-stat": "pen, notebook, stationery",
	"shop-other": "shopping, purchase",
	// Financial
	"fin-sal": "salary, wages, paycheck", "fin-prof": "bonus, profit, freelance, cashback",
	"fin-interest": "interest, fdr", "fin-deposit": "deposit, savings",
	"fin-with": "withdraw, cashout, atm", "fin-transfer": "transfer, bkash, nagad, rocket, send money",
	"fin-flexi": "recharge, flexi, flexiload, topup, data", "fin-ccpay": "credit card, cc bill",
	"fin-dps": "dps, savings scheme", "fin-loan": "bank loan, took loan",
	"fin-repay": "emi, installment, loan payment", "fin-lend": "lend, lent, gave loan, handed, covering, paid for, dhar dilam",
	"fin-recover": "recover, recovered, collected, got back, paid me back, they returned, loan repayment, ferot pelam", "fin-borrow": "borrow, borrowed, took, took loan, dhar nilam, theke nilam",
	"fin-return": "return, returned, payback, paid back, gave back, repaid, dhar shodh, ferot dilam", "fin-tax": "tax, vat, income tax",
	"fin-charge": "charge, fee, penalty", "fin-ins": "insurance, premium",
	"fin-gold": "gold", "fin-invest": "stocks, shares, crypto, mutual fund",
	"fin-misc": "financial, misc",
	// Housing
	"house-rent": "rent, house rent, bhara, basha bhara", "house-util": "electricity, gas, current bill, water bill",
	"house-net": "internet, wifi, broadband", "house-serv": "maid, bua, buya, cleaner, driver",
	"house-maint": "repair, plumbing, electrician", "house-furn": "furniture, table, chair, bed, sofa, curtain",
	"house-real": "flat, land, plot", "house-misc": "household, home",
	// Health
	"health-doc": "doctor, consultation, clinic, hospital", "health-test": "test, xray, ultrasound, lab",
	"health-med": "medicine, pharmacy, drugs", "health-other": "surgery, dental, therapy, ambulance",
	// Personal Care
	"pc-salon": "salon, haircut, shave, chuler kat, chul kata, barber, parlor", "pc-skin": "skincare, cream, facewash",
	"pc-spa": "spa, massage", "pc-toilet": "soap, shampoo, toothpaste, razor, tissue",
	"pc-fit": "gym, yoga, fitness, swimming", "pc-smoke": "cigarette, cigarettes, vape, paan",
	"pc-misc": "personal care, wellness",
	// Family
	"fam-allow": "wife, husband, spouse, pocket money", "fam-par": "parents, baba, ma, support",
	"fam-baby": "baby, diaper, formula", "fam-child": "kids, school fee, tuition, toys",
	"fam-care": "family, elder care", "fam-other": "family",
	// Education
	"edu-course": "course, coaching, training, tuition", "edu-book": "books, book, textbook, study materials",
	"edu-exam": "exam, exam fee, admission", "edu-other": "education",
	// Entertainment
	"ent-movie": "movie, cinema", "ent-sub": "netflix, youtube, spotify, subscription",
	"ent-rec": "park, zoo, outing, picnic, hangout", "ent-game": "games, game, gaming",
	"ent-event": "concert, event, ticket", "ent-misc": "hobby, entertainment",
	// Travel
	"trv-ticket": "flight, ticket, travel", "trv-hotel": "hotel, resort, airbnb",
	"trv-dine": "travel food, dining", "trv-sight": "sightseeing, tour, attraction",
	"trv-trans": "travel transport, rental", "trv-gift": "souvenir, travel gift",
	"trv-misc": "trip, journey, travel",
	// Festival
	"fest-eid": "eid, eidi, salami", "fest-wed": "wedding, biye, gift",
	"fest-others": "puja, christmas, boishakh", "fest-gift": "festival gift, birthday gift",
	"fest-decor": "decoration, lights, flowers", "fest-charity": "zakat, fitra, donation, sadaqah",
	"fest-food": "feast, dawat, iftar, sehri",
	// Miscellaneous
	"misc-init": "opening balance, initial", "misc-gift": "gift",
	"misc-charity": "charity, tip, help", "misc-office": "office, work expense",
	"misc-loss": "lost, stolen, theft", "misc-adj": "adjustment, correction",
	"misc-misc": "other, misc",
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

	assignSubcategoryTypes()

	catIDToName := make(map[string]string)
	for _, cat := range TxnCategories {
		catIDToName[cat.ID] = cat.Name
	}

	catTypeSet := make(map[string]map[TransactionType]bool)
	for i := range TxnSubcategories {
		// Mutate in place so Keywords is carried into the taxonomy served to the
		// UI, the AI classifier (marshaled from this slice), and SubcategoryByID.
		TxnSubcategories[i].Keywords = subKeywords[TxnSubcategories[i].ID]
		sub := TxnSubcategories[i]
		SubCatNameMap[sub.ID] = sub.Name
		SubcategoryByID[sub.ID] = sub

		if catName, ok := catIDToName[sub.CatID]; ok {
			SubCatToCatNameMap[sub.ID] = catName
		} else {
			SubCatToCatNameMap[sub.ID] = "Unknown"
		}

		set, ok := catTypeSet[sub.CatID]
		if !ok {
			set = make(map[TransactionType]bool)
			catTypeSet[sub.CatID] = set
		}
		for _, t := range SubcategoryTypes[sub.ID] {
			set[t] = true
		}
	}

	for _, cat := range TxnCategories {
		CategoryTypes[cat.ID] = collectTypes(catTypeSet[cat.ID])
	}
}

// assignSubcategoryTypes populates SubcategoryTypes for every subcategory.
// Most categories are Expense-only; fin and misc are mixed and use per-subcategory maps.
func assignSubcategoryTypes() {
	expenseOnlyCats := map[string]bool{
		"food": true, "trans": true, "shop": true, "house": true,
		"health": true, "pc": true, "fam": true, "edu": true,
		"ent": true, "trv": true, "fest": true,
	}

	finSubTypes := map[string][]TransactionType{
		"fin-sal":      typIncome,
		"fin-prof":     typIncome,
		"fin-interest": typIncome,
		"fin-deposit":  typTransfer,
		"fin-with":     typTransfer,
		"fin-transfer": typTransfer,
		"fin-flexi":    typExpense,
		"fin-ccpay":    typExpense,
		"fin-dps":      typExpense,
		"fin-loan":     typIncome,
		"fin-repay":    typExpense,
		"fin-lend":     typExpense,
		"fin-recover":  typIncome,
		"fin-borrow":   typIncome,
		"fin-return":   typExpense,
		"fin-tax":      typExpense,
		"fin-charge":   typExpense,
		"fin-ins":      typExpense,
		"fin-gold":     typExpense,
		"fin-invest":   typExpense,
		"fin-misc":     typExpense,
	}

	miscSubTypes := map[string][]TransactionType{
		"misc-init":    typExpOrInc,
		"misc-gift":    typExpOrInc,
		"misc-charity": typExpOrInc,
		"misc-office":  typExpense,
		"misc-loss":    typExpense,
		"misc-adj":     typExpOrInc,
		"misc-misc":    typExpOrInc,
	}

	for _, sub := range TxnSubcategories {
		switch {
		case expenseOnlyCats[sub.CatID]:
			SubcategoryTypes[sub.ID] = typExpense
		case sub.CatID == "fin":
			SubcategoryTypes[sub.ID] = finSubTypes[sub.ID]
		case sub.CatID == "misc":
			SubcategoryTypes[sub.ID] = miscSubTypes[sub.ID]
		}
	}
}

// collectTypes returns Types in canonical order (Expense, Income, Transfer).
func collectTypes(set map[TransactionType]bool) []TransactionType {
	order := []TransactionType{ExpenseTransaction, IncomeTransaction, TransferTransaction}
	out := make([]TransactionType, 0, len(order))
	for _, t := range order {
		if set[t] {
			out = append(out, t)
		}
	}
	return out
}
