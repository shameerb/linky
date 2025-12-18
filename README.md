## Linky

A Vue.js and Go-based application for managing markdown links and notes.

### What is Linky?
Linky is a modern, lightweight application designed to solve the challenge of managing and organizing web links and notes in markdown format. It provides a clean, intuitive interface for:

- **Link Management**: Easily save, categorize, and manage web links with descriptions
- **Note Organization**: Create and organize notes in markdown format
- **Quick Access**: Instantly search and filter through your collection of links and notes
- **Local Storage**: Store all your data locally in markdown files, giving you complete control over your data
- **Portable Format**: Use standard markdown format, making it easy to view and edit your links with any text editor

Whether you're a researcher collecting references, a developer saving technical resources, or simply someone who wants to organize their bookmarks in a more structured way, Linky provides a simple yet powerful solution.

## Installation and Setup

### Prerequisites
- Go 1.21 or later
- Node.js 18 or later
- npm
- Docker (optional, for containerized deployment)

### Project Structure
```
markdown-editor-go/
├── frontend/           # Vue.js frontend application
│   ├── src/           # Vue source files
│   ├── package.json   # Frontend dependencies
│   └── vite.config.js # Vite configuration
├── backend/           # Go backend server
│   ├── main.go       # Go server code
│   ├── index.html    # Entry point HTML
│   └── static/       # Built frontend files
└── markdown/         # Default directory for markdown files
```

### Configuration

#### Environment Variables

The application supports the following environment variables:

- **`STORAGE_TYPE`** - Storage backend to use:
  - `file` (default) - File/markdown storage
  - `sqlite` - SQLite database
  - `postgres` - PostgreSQL database

- **`DB_PATH`** - Path to SQLite database file
  - Default: `./linky.db`
  - Only used when `STORAGE_TYPE=sqlite`

- **`MARKDOWN_DIR`** - Directory where markdown files are stored
  - Default: `./Links`
  - Only used when `STORAGE_TYPE=file`

- **`PORT`** - Server port
  - Default: `8080`

- **`JWT_SECRET`** - JWT secret for authentication
  - Default: `your-secret-key-change-this-in-production`
  - **Important:** Change this in production!

- **`GO_ENV`** - Environment mode
  - Set to `production` for production mode

#### Storage Backends

Linky supports multiple storage backends. Choose the one that fits your needs:

##### 1. File Storage (Default)
Uses markdown files directly. Best for personal use and easy backup.

```bash
# Uses ./Links directory by default
make run

# Or with custom directory
STORAGE_TYPE=file MARKDOWN_DIR=/path/to/links make run
```

##### 2. SQLite Storage
Uses SQLite database. Best for better performance and multi-user scenarios.

**Step 1:** Run migration to import markdown files:
```bash
go run cmd/migrate/main.go \
  -markdown-dir ./Links \
  -db-path ./linky.db \
  -email "your-email@example.com"
```

**Step 2:** Start server with SQLite:
```bash
STORAGE_TYPE=sqlite DB_PATH=./linky.db make run
```

##### 3. PostgreSQL Storage
Uses PostgreSQL database. Best for production and large-scale deployments.

```bash
STORAGE_TYPE=postgres DATABASE_URL="postgres://user:pass@host:5432/dbname" make run
```

### Running the Application

#### 1. Local Development

a. Build and run the complete application:
```bash
# Using default markdown directory (./markdown)
make run

# Using custom markdown directory
MARKDOWN_DIR=/path/to/your/markdown make run
```

b. Run frontend development server with hot-reload:
```bash
make frontend-dev
```

c. Run backend server only:
```bash
make backend-run
```

#### 2. Docker Deployment

a. Using Docker directly:
```bash
# Build the Docker image
make docker-build

# Run with default markdown directory
make docker-run

# Run with custom markdown directory
MARKDOWN_DIR=/path/to/your/markdown make docker-run
```

b. Using Docker Compose:
```bash
# Run with default markdown directory
docker-compose up

# Run with custom markdown directory
MARKDOWN_DIR=/path/to/your/markdown docker-compose up
```

### Development Commands

```bash
# Install frontend dependencies
make frontend-install

# Build frontend only
make frontend-build

# Build backend only
make backend-build

# Clean build artifacts
make clean
```

### Data Migration

Linky provides migration scripts to import your markdown links into different storage backends.

#### Migrate to SQLite Database

Import markdown files from the `Links` folder into SQLite database:

```bash
go run cmd/migrate/main.go \
  -markdown-dir ./Links \
  -db-path ./linky.db \
  -email "your-email@example.com"
```

**Parameters:**
- `-markdown-dir` - Path to markdown files (default: `./Links`)
- `-db-path` - Path to SQLite database file (default: `./linky.db`)
- `-email` - Email for user account (required)
- `-password` - Password for user account (default: `password`, optional)

**Output Example:**
```
============================================================
Migration completed successfully!
============================================================
Database: ./linky.db
User: shameer789@gmail.com
Password: password

------------------------------------------------------------
SUMMARY
------------------------------------------------------------
  ai                     5 topics    120 links
  jobs                   2 topics     45 links
  learn                  8 topics    230 links
  memfoldai              3 topics     89 links
  now                    4 topics     67 links
  python                 6 topics    145 links
  reading                2 topics     34 links
  study                  7 topics    218 links
------------------------------------------------------------
Total: 8 subjects, 37 topics, 908 links
============================================================
```

#### Migrate to File Store (JSON)

Import markdown files into JSON-based file store:

```bash
go run cmd/migrate-to-filestore/main.go \
  -markdown-dir ./Links \
  -data-dir ./data
```

**Parameters:**
- `-markdown-dir` - Path to markdown files (default: `./Links`)
- `-data-dir` - Path to file store data directory (default: `./data`)

**Note:** The file store migration uses a default user account with credentials:
- Email: `shameer789@gmail.com`
- Password: `password`

### Installing as a System Binary

For convenient access, you can install Linky as a system command:

#### 1. Build and Install

**Option A: Using Go Install (Recommended)**

This installs to `~/go/bin/linky` which is automatically in your PATH if Go is properly configured:

```bash
cd /Users/shameer/Documents/work/projects/linky
go install cmd/server/main.go

# Verify installation
which linky
```

Or add to your Makefile:
```makefile
.PHONY: install
install:
	@echo "Installing linky to ~/go/bin..."
	@go install cmd/server/main.go
	@echo "✅ linky installed successfully!"
```

Then run: `make install`

**Option B: Manual Install to /usr/local/bin**

If you prefer installing to a system directory:

```bash
# Build the binary
go build -o linky cmd/server/main.go

# Install to /usr/local/bin
sudo cp linky /usr/local/bin/

# Verify installation
which linky
```

#### 2. Create a Shell Function/Alias

##### For Fish Shell

Create `~/.config/fish/functions/run-linky.fish`:

```fish
function run-linky
    set -x STORAGE_TYPE file
    set -x MARKDOWN_DIR /Users/shameer/Dropbox/Links
    set -x DISABLE_AUTH true
    nohup linky > /dev/null 2>&1 &
    echo "linky started with:"
    echo "  STORAGE_TYPE=$STORAGE_TYPE"
    echo "  MARKDOWN_DIR=$MARKDOWN_DIR"
    echo "  DISABLE_AUTH=$DISABLE_AUTH"
end
```

Then reload: `source ~/.config/fish/config.fish`

Usage: `run-linky`

##### For Bash/Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
run-linky() {
    export STORAGE_TYPE=file
    export MARKDOWN_DIR=/Users/shameer/Dropbox/Links
    export DISABLE_AUTH=true
    nohup linky > /dev/null 2>&1 &
    echo "linky started with:"
    echo "  STORAGE_TYPE=$STORAGE_TYPE"
    echo "  MARKDOWN_DIR=$MARKDOWN_DIR"
    echo "  DISABLE_AUTH=$DISABLE_AUTH"
}
```

Then reload: `source ~/.bashrc` (or `~/.zshrc`)

Usage: `run-linky`

#### 3. For SQLite Storage

If using SQLite instead, first run the migration:

```bash
# Create data directory
mkdir -p ~/.linky

# Run migration
go run cmd/migrate/main.go \
  -markdown-dir /Users/shameer/Dropbox/Links \
  -db-path ~/.linky/linky.db \
  -email "shameer789@gmail.com"
```

Then update your shell function:
```fish
function run-linky
    set -x STORAGE_TYPE sqlite
    set -x DB_PATH ~/.linky/linky.db
    set -x DISABLE_AUTH true
    nohup linky > /dev/null 2>&1 &
    echo "linky started with:"
    echo "  STORAGE_TYPE=$STORAGE_TYPE"
    echo "  DB_PATH=$DB_PATH"
    echo "  DISABLE_AUTH=$DISABLE_AUTH"
end
```

#### Managing the Background Process

```bash
# Find the process
ps aux | grep linky

# Stop linky
pkill linky
```

### Accessing the Application

Once running, the application is available at:
- Main application: http://localhost:8080
- Frontend dev server (when using make frontend-dev): http://localhost:5173

### External Markdown Directory

You can use any external directory for your markdown files by:

1. Setting the `MARKDOWN_DIR` environment variable:
```bash
export MARKDOWN_DIR=/path/to/your/markdown
```

2. Ensuring the directory exists and has proper permissions:
```bash
mkdir -p /path/to/your/markdown
chmod 755 /path/to/your/markdown
```

3. When using Docker, the external directory must be mounted as a volume (handled automatically by the provided commands)

### Troubleshooting

1. Port 8080 already in use:
```bash
# Find the process using port 8080
lsof -i :8080

# Kill the process
kill <PID>
```

2. Permission issues with markdown directory:
- Ensure the directory exists
- Check directory permissions
- When using Docker, ensure the mounted volume has correct permissions

### Features
- Notes
- Labels/Tags links
- Link to sheets for organization management - books, courses

### Fixes
- [ ] Allow adding a new subject in add function. Default to date.
- [ ] The add function doesnt add to a subject which already exists
- [ ] Move backend data to a database instead of a file
  - [ ] add concurrency features for scalability
- [ ] supabase for auth + data online
- [ ] basic
  - [ ] the search functionality doesnt score properly because of subject categorization
  - [ ] should allow for full keyword search using ""
- [ ] labels instead of subjects. makes it easier to add and filter. can filter with multiple tags. 
  - [ ] should allow to search and add labels as well.
- [ ] Fix the makefile. Currently the build, run and install scripts are not clean
- [ ] unique identifier for links and file
- [ ] feature to archive a link once read
- [ ] quick move to subject / subject search
- [ ] Production
    - [ ] cheap domain
    - [ ] add authentication, authorization
    - [ ] readonly mode
    - [ ] logging

### **Quick Summary Checklist:**

| Category                     | Tasks                                              |
|------------------------------|----------------------------------------------------|
| **Authentication & Security**| HTTPS, JWT/OAuth2, Password Hashing, Rate Limiting |
| **Database**                 | Backup, Security, Migration                        |
| **Infrastructure**           | Docker, Cloud Hosting, CI/CD Pipeline              |
| **Logging & Monitoring**     | Structured Logs, Monitoring Tools                  |
| **Scalability**              | Load Balancing, Caching                            |
| **Compliance**               | GDPR, Privacy Policies, Audit Logging              |
| **Documentation**            | API Docs, Operational Playbooks                    |

### Enhancements
- Feature Additions
  - share view of a list of selected links to others
  - add tags for filtering
  - combine most links to a single file
  - backed by database instead of md file. (support both)
  - pagination
  - concurrency : writes will not work for multiple users properly.
  - file path validation
  - error handling in delete links. Should safeguard the file (possibly a transaction with rollback for corruption of file). Datastore seems much safer.
  - before production - logging, auth, rate limits, circuit breaker, user management, cookie management, caching etc.
  - 