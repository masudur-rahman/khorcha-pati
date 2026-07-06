# Khorcha-Pati — Self-Hosting Guide

A Telegram Bot to track your expenses. This guide covers everything you need to self-host the bot using Docker or native Go.

## Requirements

### Telegram Bot

1. Create a new bot using [BotFather](https://t.me/botfather).
    - Use `/newbot` command to create a new bot.
    - Use `/setname` command to set a name for the bot.
    - Use `/setdescription` command to set a description for the bot.

2. Set commands for the bot using `/setcommands` command.
    ```
    new - Add new Wallet or Contact
    newtxn - Add new transaction
    undo - Undo last transaction
    contacts - List contacts
    balance - List Wallet Balances
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
https://github.com/masudur-rahman/khorcha-pati/assets/13915755/bc74ec7a-b243-4faa-a07b-31ebe2260264

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

2. [Create a Service account](https://console.cloud.google.com/iam-admin/serviceaccounts/create) named `khorcha-pati` and download a service account JSON key.

3. [Enable the Google Drive API](https://console.cloud.google.com/apis/library/drive.googleapis.com) for your project.
4. On Google Drive:
    - Create a folder named `.khorcha-pati`.
    - Share this folder with the service account (`khorcha-pati@<project-id>.iam.gserviceaccount.com`) and grant it "Editor" permission.

## Environment Variables

The application supports environment variable overrides for all sensitive configuration.

### Required

| Variable | Description |
|---|---|
| `TELEGRAM_BOT_TOKEN` | Bot token from @BotFather |
| `EXPENSE_BOT_TOKEN` | (Alternative) Overrides the Telegram secret in YAML |

### AI Classification (Optional)

Set these to enable AI-powered natural language processing. The bot will automatically prefer Gemini if its key is provided.

| Variable | Description |
|---|---|
| `GEMINI_API_KEY` | Google Gemini API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |

### Database Overrides (Optional)

These override the corresponding values in the YAML config file for production/Docker environments.

| Variable | Description |
|---|---|
| `EXPENSE_DB_PASS` | Database password (Postgres) |
| `EXPENSE_REDIS_PASS` | Redis password |

### Other (Optional)

| Variable | Description |
|---|---|
| `ENV` | Set to `production` for JSON-structured logging and production Zap profile |
| `BASE_URL` | If set, the bot pings `{BASE_URL}/healthz` every 20 minutes to keep itself alive |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to Google service account JSON for Drive backup |

## Configuration File

The bot reads its configuration from `.configs/.khorcha-pati.yaml` (relative to the project root).

```yaml
telegram:
    user: <telegram_username>
database:
    type: sqlite  # or postgres
    sqlite:
        syncToDrive: false
        disableSyncFromDrive: false
    postgres:
        name: expense
        host: localhost
        port: 5432
        user: postgres
        password: "" # Use EXPENSE_DB_PASS env var to override
        sslmode: disable
cache:
    type: map  # or redis
    redis:
        host: localhost
        port: "6379"
        password: "" # Use EXPENSE_REDIS_PASS env var to override
system:
    pdfGenerator: wkhtmltopdf # or chromedp
    aiGenerator: gemini # or open-router
```

## Installation and Running

### Local Setup

1. Clone the repository:

   ```bash
   git clone git@github.com:masudur-rahman/khorcha-pati.git
   cd khorcha-pati
   ```

2. Update the configuration file `.configs/.khorcha-pati.yaml`.

3. Export required environment variables:

   ```bash
   export TELEGRAM_BOT_TOKEN=<TELEGRAM_BOT_TOKEN>
   export GEMINI_API_KEY=<YOUR_GEMINI_API_KEY>
   ```

4. Run the server:

   ```bash
   make run
   ```

### Docker Setup

The Docker image supports both `wkhtmltopdf` and `chromedp` engines and is built for both AMD64 and ARM64 (Apple Silicon).

- Write configuration file
    ```shell
    mkdir -p $HOME/.khorcha-pati/configs

    echo '
    telegram:
      user: <telegram_username>
    database:
      type: sqlite
    cache:
      type: map
    ' > $HOME/.khorcha-pati/configs/.khorcha-pati.yaml
    ```

- Run Khorcha-Pati
    ```shell
    docker run -d \
      --name khorcha-pati \
      -v $HOME/.khorcha-pati/configs:/app/.configs \
      -v $HOME/.khorcha-pati:/.khorcha-pati \
      -e TELEGRAM_BOT_TOKEN=<TELEGRAM_BOT_TOKEN> \
      -e GEMINI_API_KEY=<GEMINI_API_KEY> \
      -e ENV=production \
      ghcr.io/masudur-rahman/khorcha-pati:latest serve
    ```

### Production Environment (Kubernetes)

To deploy `Khorcha-Pati` application in production environment, the preferred way is through Helm Chart. Checkout more [here](https://github.com/masudur-rahman/helm-charts/tree/main/charts/khorcha-pati).


- First you need to add the repo for the helm chart.
    ```bash
    helm repo add masud https://masudur-rahman.github.io/helm-charts/stable
    helm repo update

    helm search repo masud/khorcha-pati
    ```
    - Install the chart
        - For installing just with SQLite database (without Google Drive backup)
          ```bash
          helm upgrade --install khorcha-pati masud/khorcha-pati -n demo \
              --create-namespace \
              --set telegram.token=<TELEGRAM_BOT_TOKEN> \
              --set telegram.user=<TELEGRAM_USERNAME>
          ```
        - SQLite with Google Drive backup
          ```bash
          helm upgrade --install khorcha-pati masud/khorcha-pati -n demo \
              --create-namespace \
              --set telegram.token=<TELEGRAM_BOT_TOKEN> \
              --set telegram.user=<TELEGRAM_USERNAME> \
              --set database.sqlite.syncToDrive=true \
              --set-file googleCredJson=<GOOGLE-SVC-ACCOUNT-JSON-FILEPATH>
          ```
        - Postgres database
          ```bash
          helm upgrade --install khorcha-pati masud/khorcha-pati -n demo \
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
  To check if `Khorcha-Pati` is installed, run the following command:
    ```bash
    $ kubectl get pods -n demo -l "app.kubernetes.io/instance=khorcha-pati"

    NAME                                            READY   STATUS    RESTARTS      AGE
    khorcha-pati-7989d96fcc-b4smq            1/1     Running   2 (30s ago)   31s
    khorcha-pati-postgres-55dcb67965-95r7g   1/1     Running   0             31s
    ```

## Health Check

The bot exposes a health check endpoint:

- **Endpoint:** `GET /healthz` on port `8080`
- **Response:** JSON with database connectivity status
- **Docker:** The Docker image includes a built-in `HEALTHCHECK` directive

If `BASE_URL` is set, the bot will ping `{BASE_URL}/healthz` every 20 minutes as a keep-alive mechanism.
