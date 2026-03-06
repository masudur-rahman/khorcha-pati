# Expense Tracker

A Telegram Bot to track your expenses.
`Expense Tracker Bot` is a Telegram Bot to track your daily transactions.

## Features

- **Expense Tracking**: Keep track of your daily expenses, income, and balance transfers between accounts.
- **Flexible Input**: Add transactions interactively by selecting options or simply send a text describing your transaction.
- **Lending and Borrowing**: Track lendings and borrowings with other individuals.
- **Transaction Summary**: Retrieve transaction summaries based on type, category, or subcategory for your preferred duration.
- **Transaction Reports**: Generate transaction reports in PDF format for your chosen duration.

## Usage
The `Expense Tracker Bot` is now available for public use.
To use this bot, go to Telegram and search for [@XpenseTrackerBot](https://t.me/XpenseTrackerBot)

Once you are inside the bot inbox,  press `Start` button to start using the Tracker Bot.

Before you start tracking your expenses
- Add wallets like `cash`, `brac`, `ebl` etc
  - Command `/new` => Wallet => Type (Cash or Bank)
  - Reply with wallet details (`cash "Cash in Hand"`, `ebl EBL` etc)
- Add contacts with whom you are financially involved
  - Command `/new` => Contact
  - Reply with the contact details (`john "John Doe" john@doe.com`)

### Track your Transactions

#### Interactively
To track your transactions interactively, send `/newtxn` command and follow the on-display suggestions.

#### Regular Text Message (Natural Language)
You can add new transactions by simply sending a text message describing what you did. The bot is smart enough to understand natural language!

You just need to mention:
- **What** you did (description/category)
- **How much** (amount)
- **When** (optional, defaults to now)
- **Wallet** (optional, defaults to 'cash')
- **Contact** (optional, for loans/borrows)

**Examples:**
```
lunch 250
groceries 1.5k
bought a new shirt for 2500
transfer 10k from brac to city
lent 5000 to karim
got bonus 20k
dinner 1500 yesterday
internet 500 on 1st
spent 500 for taxi
```

<details>
<summary>Expand to see the Allowed Transaction Subcategory list</summary>

```
Food (food):
- Grocery (food-groc)
- Vegetables (food-veg)
- Fruits (food-fruit)
- Fish (food-fish)
- Meat (food-meat)
- Dairy & Eggs (food-dairy)
- Bakery (food-bakery)
- Restaurant (food-rest)
- Street Food (food-street)
- Takeout (food-take)
- Snack/Tea (food-snack)
- Beverages (food-bev)
- General (food-misc)

Transport (trans):
- Bus/Train (trans-pub)
- Taxi/Ride (trans-taxi)
- Fuel (trans-fuel)
- Tolls/Parking (trans-toll)
- Vehicle Maint (trans-maint)
- Other Transport (trans-other)

Shopping (shop):
- Household Supplies (shop-supply)
- Clothing (shop-cloth)
- Footwear (shop-foot)
- Electronics (shop-elec)
- Jewelry (shop-jewelry)
- Cosmetics (shop-beauty)
- Accessories (shop-acc)
- Stationary (shop-stat)
- Gifts (Purchased) (shop-gift)
- General Shopping (shop-other)

Financial (fin):
- Salary (fin-sal)
- Profit/Bonus (fin-prof)
- Interest (fin-interest)
- Deposit (fin-deposit)
- Withdraw (fin-withdraw)
- Acc Transfer (fin-transfer)
- Mobile Recharge (fin-flexi)
- Credit Card Payment (fin-ccpay)
- DPS (fin-dps)
- Bank Loan (Taken) (fin-loan-get)
- Bank Repayment (fin-repay)
- Lending (Given) (fin-lend)
- Lend Recovery (Received) (fin-recover)
- Borrowing (Taken) (fin-borrow)
- Borrow Return (Given) (fin-return)
- VAT/Tax (fin-tax)
- Charges (fin-charge)
- Insurance (fin-ins)
- Gold Investment (fin-gold)
- Stocks/Assets (fin-invest)
- Overhead (fin-misc)

Housing (house):
- Rent (house-rent)
- Utilities (house-util)
- Internet (house-net)
- Maid/Service (house-serv)
- Maintenance (house-maint)
- Furniture (house-furn)
- Real Estate (house-real)
- General Household (house-misc)

Health (health):
- Doctor Visit (health-doc)
- Medical Tests (health-test)
- Medicine (health-med)
- Other Health Exp (health-other)

Personal Care (pc):
- Salon (pc-salon)
- Skincare (pc-skin)
- Spa & Massage (pc-spa)
- Toiletries (pc-toilet)
- Fitness (pc-fit)
- Wellness (pc-misc)

Family (fam):
- Spouse Allowance (fam-allow)
- Parents (fam-par)
- Baby Needs (fam-baby)
- Kids Needs (fam-child)
- Family Care (fam-care)
- Other Family Exp (fam-other)

Education (edu):
- Courses (edu-course)
- Books/Stationary (edu-book)
- Exam Fees (edu-exam)
- Other Education (edu-other)

Entertainment (ent):
- Movies (ent-movie)
- Subscription (ent-sub)
- Recreation (ent-rec)
- Gaming (ent-game)
- Concerts/Events (ent-event)
- Hobby/Misc (ent-misc)

Travel (trv):
- Tickets (trv-ticket)
- Hotel (trv-hotel)
- Dining (trv-dine)
- Sightseeing (trv-sight)
- Transportation (trv-trans)
- Gifts (trv-gift)
- Journey (trv-misc)

Festival (fest):
- Eid (fest-eid)
- Wedding (fest-wedding)
- Other Festivals (fest-others)
- Gifts (fest-gift)
- Decoration (fest-decor)
- Zakat/Donation (fest-charity)
- Fest Feast (fest-food)

Miscellaneous (misc):
- Initial Amount (misc-initial)
- General Gifts (misc-gift)
- General Charity (misc-charity)
- Office/Work Exp (misc-office)
- Lost/Stolen (misc-loss)
- Balance Adjustment (misc-adj)
- General (misc-misc)
```

</details>

You always can send `/cat` command to list the subcategory

### Available commands:
- `/new` - Add new Wallet or Contact
  - `Wallet` - Add new wallet (Cash, Bank)
    - i.e: `brac "BRAC Bank"`
    - i.e: `cash "Cash in Hand"`
  - `Contact` - Add a person for lending/borrowing

- `/newtxn` - Add new transaction (Interactive)
- `/undo` - Undo last transaction (soft-delete + revert balances)
- `/contacts` - List contacts
  - list all the people involved in lending/borrowing with you
- `/balance` - List Wallet Balances
  - list all the registered wallets and their balance
- `/list` - List recent transactions
- `/expense` - Fetch Expense of Current month
  - list transactions of current month
- `/summary` - Transaction summary of current month
  - list transaction summary of current month
- `/allsummary` - Transaction summary based on Type, Category, Subcategory
  - list transaction summary based on Type, Category, Subcategory
  - with a duration query parameter
- `/report` - Transaction Report
  - list transaction report
  - with a duration query parameter
- `/cat` - List Transaction categories
  - list all the registered categories
  - by selecting a category, list all the registered subcategories of that category
- `/sync` - Sync database to Google Drive (if configured)
- `/help` - Show Usage page


## Live Demonstration

https://github.com/masudur-rahman/expense-tracker-bot/assets/13915755/83db45c8-1e84-473e-8d58-cda6ef8cc6ef

## Future Work

A list of possible future work:
- [x] Add support for undoing a transaction
- [x] Add Database backup and restore support (Google Drive sync)
- [ ] Add support for multiple users

## Self Hosting

If you want to host your own `Expense Tracker Bot`, refer to the [self-hosting](./docs-selfhost) doc page.

## Contacts

Telegram - [masudur-rahman](https://t.me/masudur_rahman).
