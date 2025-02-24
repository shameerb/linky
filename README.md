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
The application requires one environment variable:
- `MARKDOWN_DIR`: Directory where markdown files are stored
  - Default: `./markdown` in the project root
  - Can be set to any accessible directory path

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

### Enhancements
#### Immediate
- way to use keyboards to move cursor, open tabs, select links, delete
- Shortcuts
    - x - select
    - shift + 3 (#) - delete
- copy links from current window to the page with timestamp as subject
- add links in bulk
- unique identifier for links and file
- saved links into a datastore or file (factory pattern)
    - move the data to a db instead of file
- feature to archive a link once read
- view to show only a limited links per page if necessary. (analysis paralysis)
- change to vuejs app
- Public url
    - deploy to a public url
    - authentication
- Feature Additions
    - ability to add bulk links
    - add tags for filtering
    - select only 1 subject
    - combine most links to a single file
    - backed by database instead of md file. (support both)
    - books list
        - only categories and links. Sourced from where? how to read from browser on mobile ?
    - pagination