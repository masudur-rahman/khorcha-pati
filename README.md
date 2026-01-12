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
- Add accounts like `cash`, `brac`, `ebl` etc
  - Command `/new` => Account => Type (Cash or Bank)
  - Reply with account details (`cash "Cash in Hand"`, `ebl EBL` etc)
- Add some debtors/creditors with whom you are financially involved
  - Command `/new` => DebtorsCreditors
  - Reply with the person details (`john "John Doe" john@doe.com`)

### Track your Transactions

#### Interactively
To track your transactions interactively, send `/newtxn` command and follow the on-display suggestions.

#### Regular Text Message
You also can add new transaction by sending a regular text message.  
You just need to mention
- what you did
- how much did it cost
- when you did it
- affected accounts
- affected persons in case of loan/borrow
- remarks

and the bot will take care of the rest

##### Obviously you need to follow some rules while adding transactions via text messages.

Message needs to be in key/value pairs, like:
- {action} {amount}
- for {txn subcategory} // the subcategory must match the allowed subcategory
- {from/to} {account} // default cash
- {from/to} {debtor/creditor} // in case of lending/borrowing
- on {date} // "DD-MM-YYY", "YYYY-MM-DD", "MMM DD, YYYY", today, tomorrow, yesterday [default today]
- at {time} // midnight, morning, noon, afternoon, evening, night and also different time formats [default now] 
- note {remarks}

These key/value pairs can appear in any order


Some example text for adding a new transaction:
```
- transferred 2000 from brac to dbbl on 2020-01-01 note "Bill payment"
- spent 1000 for food-rest on "Jan 13, 2013" from dbbl note "Lunch"
- earn 5000 to brac on 20-01-2023 note "Salary"
- borrow 1000 from user to brac on 2020-01-01
- return 1000 to user from brac on 2020-01-01
- lend 1000 to user from brac on 2020-01-01
- recover 1000 from user to brac on 2020-01-01
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
- `/new` - Add new Transaction, Account or User
  - `DebtorsCreditors` - Add new debtor or creditors
  - `Account` - Add new account (Cash, Bank)
    - i.e: `brac "BRAC Bank"`
    - i.e: `cash "Cash in Hand"`

[//]: # (  - `txn` - Add new transaction)
[//]: # (    - through some flags)
[//]: # (    - i.e: `<amount> -t=<type> -s=<subcat> -f=<src> -d=<dst> -u=<user> -r=<remarks>`)

- `/newtxn` - Add new transaction
  - Add a new transaction through a series of callback queries.
- `/users` - List users
  - list all the persons involved in some loan/borrow with the system user
- `/balance` - List Account Balance
  - list all the registered accounts and their balance
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
- `/help` - Show Usage page


## Live Demonstration

https://github.com/masudur-rahman/expense-tracker-bot/assets/13915755/83db45c8-1e84-473e-8d58-cda6ef8cc6ef

## Future Work

A list of possible future work:
- [ ] Add support for undoing a transaction
- [ ] Add Database backup and restore support
- [ ] Add support for multiple users

## Self Hosting

If you want to host your own `Expense Tracker Bot`, refer to the [self-hosting](./docs-selfhost) doc page.

## Contacts

Telegram - [masudur-rahman](https://t.me/masudur_rahman).
