# Expense Tracker

A Telegram Bot to track your expenses.

## Description

`Expense Tracker Bot` is a Telegram Bot to track your expenses. It is built using [Go](https://golang.org/).

## Requirements

### Telegram Bot

1. Create a new bot using [BotFather](https://t.me/botfather).
    - Use `/newbot` command to create a new bot.
    - Use `/setname` command to set a name for the bot.
    - Use `/setdescription` command to set a description for the bot.

2. Set commands for the bot using `/setcommands` command.
    ```
    new - Add new Transaction, Account or User
    newtxn - Add new transaction
    users - List persons involved in some loan/borrow with the system user
    balance - List Account Balance
    list - List recent transactions
    expense - Fetch Expense of Current month
    summary - Transaction summary of current month
    allsummary - Transaction summary based on Type, Category, Subcategory
    report - Transaction Report
    cat - List Transaction categories
    sync - Sync database to Google Drive
    help - Show Usage page
    ```

3. Create a Token for the bot.
    - Use `/token` command to get the bot token.

#### Telegram Bot Creation Demo
https://github.com/masudur-rahman/expense-tracker-bot/assets/13915755/bc74ec7a-b243-4faa-a07b-31ebe2260264

### Database Setup

#### SQLite (Default)

By default, the application uses SQLite as its database, requiring no additional setup.

#### PostgreSQL (Optional)

If you prefer to use PostgreSQL, follow these steps:

##### Local Setup

1. Install PostgreSQL using [Homebrew](https://brew.sh/):
   ```bash
   brew update
   brew install postgresql
   brew services start postgresql
   ```

2. Create a superuser named `postgres` with password `postgres`:
   ```bash
   psql postgres -c "CREATE USER postgres WITH SUPERUSER PASSWORD 'postgres';"
   ```

3. Create a new database named `expense`:
   ```bash
   psql -u postgres -c "CREATE DATABASE expense;"
   ```

## Google Drive Access (Optional)

If you want to back up your SQLite database to Google Drive regularly, follow these steps:

1. [Create a Google Project](https://console.cloud.google.com/projectcreate) (if not already created).

2. [Create a Service account](https://console.cloud.google.com/iam-admin/serviceaccounts/create) named `expense-tracker` and download a service account JSON key.

3. [Enable the Google Drive API](https://console.cloud.google.com/apis/library/drive.googleapis.com) for your project.
4. On Google Drive:
    - Create a folder named `.expense-tracker`.
    - Share this folder with the service account (`expense-tracker@<project-id>.iam.gserviceaccount.com`) and grant it "Editor" permission.

## Installation and Running

### Local Setup

1. Clone the repository:

   ```bash
   mkdir -p $GOPATH/src/github.com/masudur-rahman
   cd $GOPATH/src/github.com/masudur-rahman
   git clone git@github.com:masudur-rahman/expense-tracker-bot.git
   ```

2. Update the configuration file (`configs/.expense-tracker.yaml`) as needed. You can modify the Telegram user and specify the database type.
    ```bash
    telegram:
      user: masudur_rahman
    database:
      type: sqlite
    ```

3. Export required environment variables:

   ```bash
   export TELEGRAM_BOT_TOKEN=<TELEGRAM_BOT_TOKEN>
   ```

   If backing up to Google Drive:

   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS=$HOME/Downloads/service-account-key.json
   ```

4. Run the server:

   ```bash
   make run
   ```
### Docker Setup
- Write configuration file
    ```shell
    mkdir -p $HOME/.expense-tracker/configs

    echo '
    telegram:
      user: masudur_rahman
    database:
      type: sqlite
    ' > $HOME/.expense-tracker/configs/.expense-tracker.yaml
    ```

- Run Expense Tracker Bot
    ```shell
    docker run -v $HOME/.expense-tracker/configs:/configs \
      -v $HOME/.expense-tracker:/.expense-tracker \
      -e TELEGRAM_BOT_TOKEN=<TELEGRAM_BOT_TOKEN> \
      ghcr.io/masudur-rahman/expense-tracker-bot:v1.0.0 serve
    ```

### Back4App Setup

1. Create a new app on [Back4App](https://www.back4app.com/).
2. Connect your app to a GitHub repository.
3. Set the environment variables to the `Settings.Environment Variables` section.
4. Restart the server.

### Production Environment (Kubernetes)

To deploy `Expense Tracker Bot` application in production environment, the preferred way is through Helm Chart. Checkout more [here](https://github.com/masudur-rahman/helm-charts/tree/main/charts/expense-tracker-bot).


- First you need to add the repo for the helm chart.
    ```bash
    helm repo add masud https://masudur-rahman.github.io/helm-charts/stable
    helm repo update

    helm search repo masud/expense-tracker-bot
    ```
    - Install the chart
        - For installing just with SQLite database (without Google Drive backup)
          ```bash
          helm upgrade --install expense-tracker-bot masud/expense-tracker-bot -n demo \
              --create-namespace \
              --set telegram.token=<TELEGRAM_BOT_TOKEN> \
              --set telegram.user=<TELEGRAM_USERNAME>
          ```
        - SQLite with Google Drive backup
          ```bash
          helm upgrade --install expense-tracker-bot masud/expense-tracker-bot -n demo \
              --create-namespace \
              --set telegram.token=<TELEGRAM_BOT_TOKEN> \
              --set telegram.user=<TELEGRAM_USERNAME> \
              --set database.sqlite.syncToDrive=true \
              --set-file googleCredJson=<GOOGLE-SVC-ACCOUNT-JSON-FILEPATH>
          ```
        - Postgres database
          ```bash
          helm upgrade --install expense-tracker-bot masud/expense-tracker-bot -n demo \
              --create-namespace \
              --set telegram.token=<TELEGRAM_BOT_TOKEN> \
              --set telegram.user=<TELEGRAM_USERNAME> \
              --set database.type=postgres \
              --set database.deploy=true # set to false if you want to use external database
              # --set database.postgres.user=<POSTGRES_USER> \
              # --set database.postgres.password=<POSTGRES_PASSWORD> \
              # --set database.postgres.db=<POSTGRES_DB> \
              # --set database.postgres.host=<POSTGRES_HOST> \
              # --set database.postgres.port=<POSTGRES_PORT> \
              # --set database.postgres.sslmode=<POSTGRES_SSL_MODE>
          ```
- Verify Installation
  To check if `Expense Tracker Bot` is installed, run the following command:
    ```bash
    $ kubectl get pods -n demo -l "app.kubernetes.io/instance=expense-tracker-bot"

    NAME                                            READY   STATUS    RESTARTS      AGE
    expense-tracker-bot-7989d96fcc-b4smq            1/1     Running   2 (30s ago)   31s
    expense-tracker-bot-postgres-55dcb67965-95r7g   1/1     Running   0             31s
    ```
