# Expense Tracker Backend (Gin + MariaDB)

This is a starter backend project using:

- Go
- Gin (HTTP framework)
- MariaDB (via `database/sql`)
- Layered structure (`handler -> service -> repository -> db`)

## Project Structure

```txt
backend/
  cmd/api/main.go                # app entrypoint
  cmd/bot/main.go                # Discord bot entrypoint
  internal/
    config/                      # env/config loading
    db/                          # database connection setup
    discordbot/                  # Discord slash command handlers
    transactionapi/              # HTTP client used by the Discord bot
    handler/                     # HTTP request/response logic
    model/                       # domain structs
    repository/                  # SQL queries / DB access
    router/                      # route registration
    service/                     # business logic
  migrations/
    001_create_transactions.sql  # create DB + transactions table
  scripts/
    run-migration.ps1            # Windows migration helper
    run-migration.sh             # Linux/Mac migration helper
  .env.example
```

## Quick Start

## 1) Install dependencies

```bash
cd backend
go mod tidy
```

## 2) Setup environment

```bash
cp .env.example .env
```

If you are on PowerShell:

```powershell
Copy-Item .env.example .env
```

Edit `.env` with your MariaDB credentials.

## 3) Create DB and table

Option A: using mysql client directly

```bash
mysql -u root -p < migrations/001_create_transactions.sql
```

Option B (PowerShell helper):

```powershell
./scripts/run-migration.ps1 -User root -Password your_password
```

## 4) Run API server

```bash
go run ./cmd/api
```

Server default: `http://localhost:8080`

## 5) Run Discord bot

Set these values in `.env`:

```txt
DISCORD_TOKEN=your_discord_bot_token
DISCORD_GUILD_ID=your_test_server_id
TRANSACTION_API_BASE_URL=http://127.0.0.1:8080/api/v1
```

`DISCORD_GUILD_ID` is optional, but useful during development because guild commands appear quickly. If it is empty, commands are registered globally and may take longer to show up.

Start the API first, then run the bot:

```bash
go run ./cmd/bot
```

Available slash commands:

- `/transaction add`
- `/transaction list`
- `/transaction get`
- `/transaction update`
- `/transaction delete`

## API Endpoints

- `GET /health`
- `GET /api/v1/transactions?limit=20&offset=0`
- `GET /api/v1/transactions/:id`
- `POST /api/v1/transactions`
- `PUT /api/v1/transactions/:id`
- `DELETE /api/v1/transactions/:id`

### Example Create Transaction

```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Lunch",
    "amount": 7.50,
    "category": "food",
    "notes": "fried rice",
    "transaction_date": "2026-05-27",
    "type": "E"
  }'
```

## How To Add New Functionality (Beginner Guide)

Follow this order for each new feature:

## 1) Define or update model

File: `internal/model/`

Example: add a new field to `Transaction` struct.

## 2) Add repository method

File: `internal/repository/transaction_repository.go`

- Add SQL query function (read/write DB).
- Keep repository focused on DB access only.

## 3) Add service method

File: `internal/service/transaction_service.go`

- Add validation and business rules.
- Service calls repository methods.

## 4) Add handler endpoint

File: `internal/handler/transaction_handler.go`

- Parse request (`JSON`, query, path params).
- Call service method.
- Return HTTP status + JSON response.

## 5) Register route

File: `internal/router/router.go`

- Add a route and map it to your handler method.

## 6) (If needed) update migration

Folder: `migrations/`

- Add new migration SQL file.
- Do not edit old migration in real projects; create a new one.

## Example: Add "monthly summary" endpoint

1. Repository: write SQL `SUM(amount)` grouped by month.
2. Service: process/validate month/year input.
3. Handler: create `GET /api/v1/reports/monthly`.
4. Router: register route.

## Notes

- Keep function responsibilities small.
- Put HTTP-only logic in handler, not service/repository.
- Put SQL only in repository.
- Put business rules in service.
