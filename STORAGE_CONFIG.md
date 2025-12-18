# Storage Backend Configuration

Linky supports multiple storage backends using the **Repository Pattern**. You can switch between different backends without changing the frontend or API.

## Available Storage Backends

1. **SQLite** (default) - Embedded database, multi-user support
2. **File-based** - Markdown files, legacy mode
3. **PostgreSQL** - Production-grade database (coming soon)

## Configuration

Use environment variables to configure the storage backend:

### Option 1: SQLite (Default)

```bash
# Uses SQLite database with multi-user support
export STORAGE_TYPE=sqlite
export DB_PATH=./linky.db
export JWT_SECRET=your-secret-key

PORT=8081 go run cmd/server/main.go
```

### Option 2: File-based (JSON with Default User)

```bash
# Uses JSON files with a default user (no real authentication)
export STORAGE_TYPE=file
export MARKDOWN_DIR=./Links        # For legacy markdown API
export FILE_DATA_DIR=./data         # For storing subjects/topics/links as JSON

PORT=8081 go run cmd/server/main.go
```

**Default credentials for file-based storage:**
- Email: `default@local`
- Password: `password`

**Note:** File-based storage uses a single default user. All data is stored in `./data/linky-data.json`.

### Option 3: PostgreSQL (Coming Soon)

```bash
# Uses PostgreSQL database
export STORAGE_TYPE=postgres
export DATABASE_URL=postgres://user:password@localhost:5432/linky
export JWT_SECRET=your-secret-key

PORT=8081 go run cmd/server/main.go
```

## Environment Variables Reference

| Variable | Description | Default | Used By |
|----------|-------------|---------|---------|
| `STORAGE_TYPE` | Storage backend type: `sqlite`, `file`, or `postgres` | `sqlite` | All |
| `MARKDOWN_DIR` | Directory containing markdown files | `./Links` | File-based |
| `DB_PATH` | Path to SQLite database file | `./linky.db` | SQLite |
| `DATABASE_URL` | PostgreSQL connection string | - | PostgreSQL |
| `JWT_SECRET` | Secret key for JWT tokens | (default test key) | SQLite, PostgreSQL |
| `PORT` | Server port | `8080` | All |
| `GO_ENV` | Environment: `production` or development | - | All |

## Switching Between Backends

### Development: Test with different backends

```bash
# Test with SQLite
STORAGE_TYPE=sqlite PORT=8081 go run cmd/server/main.go

# Test with file-based
STORAGE_TYPE=file PORT=8082 go run cmd/server/main.go
```

### Production: Use Docker or systemd

**Docker example:**
```dockerfile
ENV STORAGE_TYPE=postgres
ENV DATABASE_URL=postgres://user:pass@db:5432/linky
ENV JWT_SECRET=secure-random-secret
```

**systemd service:**
```ini
[Service]
Environment="STORAGE_TYPE=sqlite"
Environment="DB_PATH=/var/lib/linky/linky.db"
Environment="JWT_SECRET=your-secret-key"
Environment="PORT=8080"
ExecStart=/usr/local/bin/linky
```

## Migration Between Backends

### From Files to SQLite

```bash
# 1. Run migration script
go run cmd/migrate/main.go \
  -markdown-dir ./Links \
  -db-path ./linky.db \
  -email your@email.com \
  -password yourpassword

# 2. Start server with SQLite
STORAGE_TYPE=sqlite PORT=8081 go run cmd/server/main.go
```

### From SQLite to PostgreSQL (Coming Soon)

```bash
# Export from SQLite
sqlite3 linky.db .dump > backup.sql

# Import to PostgreSQL
psql -U user -d linky < backup.sql
```

## Architecture

### Repository Pattern

```
┌─────────────────────────────────────┐
│         API Handlers                │
│  (Internal/api/handlers.go)         │
└─────────────┬───────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│      Store Interface                │
│  - Store (legacy file API)          │
│  - UserStore (database API)         │
└─────────────┬───────────────────────┘
              │
     ┌────────┴─────────┬──────────┐
     ▼                  ▼          ▼
┌──────────┐    ┌─────────────┐  ┌──────────┐
│ Markdown │    │   SQLite    │  │ Postgres │
│  Store   │    │   Store     │  │  Store   │
└──────────┘    └─────────────┘  └──────────┘
```

### Store Factory

The `store.NewStoreFromEnv()` factory function:
1. Reads `STORAGE_TYPE` environment variable
2. Initializes the appropriate backend
3. Returns both `Store` and `UserStore` interfaces
4. API handlers work with interfaces, not concrete types

## Adding a New Backend

To add a new storage backend (e.g., MongoDB):

1. **Create implementation:**
   ```go
   // internal/services/store/mongodb/mongodb.go
   type MongoStore struct {
       // ...
   }

   func (m *MongoStore) CreateUser(...) (*models.User, error) {
       // Implementation
   }
   // ... implement all UserStore methods
   ```

2. **Add to factory:**
   ```go
   // internal/services/store/factory.go
   case StorageTypeMongo:
       return newMongoStore(config)
   ```

3. **Use it:**
   ```bash
   STORAGE_TYPE=mongo MONGO_URL=mongodb://localhost:27017/linky go run cmd/server/main.go
   ```

## Best Practices

1. **Development:** Use SQLite or file-based
2. **Production:** Use PostgreSQL or managed database
3. **Testing:** Use in-memory SQLite (`:memory:`)
4. **Backup:** Regular database dumps/exports
5. **Secrets:** Never commit `JWT_SECRET` or database credentials

## Troubleshooting

### "Storage type not supported"
- Check `STORAGE_TYPE` is one of: `file`, `sqlite`, `postgres`
- Check for typos in environment variable

### "Failed to initialize store"
- For SQLite: Check `DB_PATH` directory exists and is writable
- For file: Check `MARKDOWN_DIR` exists and has `.md` files
- For PostgreSQL: Check `DATABASE_URL` and database connectivity

### Migration issues
- Ensure source files/database exist
- Check permissions
- Verify schema compatibility
