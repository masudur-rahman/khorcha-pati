# Khorcha-Pati

[![Latest release](https://img.shields.io/github/v/release/masudur-rahman/khorcha-pati?sort=semver&color=2684FF)](https://github.com/masudur-rahman/khorcha-pati/releases/latest)

A Telegram Bot — now with a web dashboard — to track your expenses.
`Khorcha-Pati` is a Telegram Bot to track your daily transactions.

## Features

- **Web Dashboard**: A beautiful, mobile-friendly React dashboard to visualize your spending trends, manage budgets, generate custom statements, and configure user settings.
- **Natural Language Input**: The bot natively understands natural language phrases (e.g., "add 500", "lunch 250"). Just text it what you spent!
- **AI-Powered Categorization**: Natural-language transactions are automatically sorted into the right category and subcategory using AI (Google Gemini and/or OpenRouter, with rotation and failover), and results are cached so repeats are instant.
- **Flexible Tracking**: Keep track of your daily expenses, income, and balance transfers between accounts.
- **Lending and Borrowing**: Track lendings and borrowings with other individuals.
- **Paginated Lists**: View your transactions and expenses in clean, paginated Markdown lists with easy navigation.
- **Hierarchical Summaries**: Retrieve transaction summaries formatted with beautiful hierarchical tree connectors for better readability on mobile.
- **Transaction Reports**: Download detailed PDF reports directly from the Telegram bot or generate interactive custom statements via the web dashboard.

## Premium User Experience

- **Auto Keyboard Cleanup**: Automatically removes sticky `ForceReply` or `ReplyKeyboardMarkup` states when switching between commands.
- **Smart NLP**: Understands natural phrases like `add 500` or `plus 1000` to quickly initialize or adjust your balances.
- **Fast Pagination**: Command responses like `/list` and `/expense` use stateful pagination to keep your chat clean and responsive.
- **Mobile-First Formatting**: Summaries and lists are formatted with hierarchical tree connectors for perfect readability on mobile screens.

## Usage
The `Khorcha-Pati` bot is now available for public use.
To use this bot, go to Telegram and search for [@KhorchaPatiBot](https://t.me/KhorchaPatiBot)

Once you are inside the bot inbox, press `Start` button to start using the Tracker Bot.

A `cash` wallet is created for you automatically — you can start tracking right away.

Optionally, before you start:
- Add bank wallets like `brac`, `ebl` etc
  - Command `/new` => Wallet => Bank
  - Reply with wallet details (`brac BRAC Bank`, `ebl EBL` etc)
- Add contacts with whom you are financially involved
  - Command `/new` => Contact
  - Reply with the contact details (`john John Doe john@doe.com`)

### Track your Transactions

#### Interactively
To track your transactions interactively, send `/newtxn` command and follow the on-display suggestions.

#### Regular Text Message (Natural Language)
You can add new transactions by simply sending a text message describing what you did. The bot is smart enough to understand natural language and will automatically categorize your expenses!

You just need to mention:
- **What** you did (description/category)
- **How much** (amount)
- **When** (optional, defaults to now)
- **Wallet** (optional, defaults to 'cash')
- **Contact** (optional, for loans/borrows)

**Examples:**
```
add 500
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
- Bank Loan (Taken) (fin-loan)
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
- Initial Amount (misc-init)
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
- `/new` - Add new wallet or contact
- `/newtxn` - Add new transaction (interactive)
- `/balance` - Show wallet balances
- `/contacts` - List contacts and balances
- `/list` - List recent transactions (paginated)
- `/expense` - List expenses (paginated)
- `/summary` - Monthly transaction summary
- `/allsummary` - Detailed summary by type, category, or subcategory
- `/report` - Generate PDF transaction report
- `/budget` - View and manage spending budgets
- `/cat` - Browse transaction categories
- `/undo` - Undo last transaction
- `/dashboard` - Open web dashboard (one-tap login)
- `/phone` - Share phone number for dashboard login
- `/sync` - Sync database to Google Drive (if configured)
- `/help` - Show usage help

### Web Dashboard

The easiest way in: send `/dashboard` to the bot and tap the button — you're signed in automatically.

From the dashboard's sign-in page you can also use:
- **QR Scan** — scan with your phone and approve in Telegram
- **OTP Code** — enter your Telegram username or phone number; the bot sends you a login code. Country code on the phone number is optional.

No Telegram username? Share your number once with `/phone`, then log in with it.


## Live Demonstration

https://github.com/masudur-rahman/khorcha-pati/assets/13915755/83db45c8-1e84-473e-8d58-cda6ef8cc6ef

## Core Stack

- **Backend**: Go (Golang), Telebot, Styx ORM
- **Database**: SQLite & PostgreSQL supported
- **Frontend**: React, TypeScript, Vite
- **AI Integration**: Google Gemini & OpenRouter APIs
- **Deployment**: Docker (Multi-arch), Kubernetes, Terraform

## Self Hosting

If you want to host your own `Khorcha-Pati` bot, refer to the [self-hosting](./docs-selfhost) doc page.

## Contacts

Telegram - [masudur-rahman](https://t.me/masudur_rahman).