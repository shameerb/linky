# Linky Migration Guide

This guide explains the new multi-user authentication system and how to migrate your existing data.

## What's New

### 1. Multi-User Support
- Each user now has their own account
- Users can only see and manage their own links
- Secure authentication with JWT tokens

### 2. New Data Structure
- **Users**: Each person has their own account
- **Subjects**: Top-level categories (previously files like "now.md", "learn.md")
- **Topics**: Sub-categories within subjects (previously ### headings in markdown)
- **Links**: Your bookmarked URLs
- **Tags**: Additional categorization for links (new feature!)

### 3. Improved UI
- Full-width layout using entire screen
- Compact link display
- Better search with OR/AND operators
- Global keyboard navigation
- Modern authentication pages

## Migration Steps

### Step 1: Build the Application

```bash
make build
```

### Step 2: Run the Migration Script

The migration script will:
- Create a SQLite database
- Import all your existing markdown files
- Create a user account for you

Run the migration:

```bash
go run cmd/migrate/main.go \
  -markdown-dir ./Links \
  -db-path ./linky.db \
  -email your@email.com \
  -password yourpassword
```

**Example:**
```bash
go run cmd/migrate/main.go \
  -markdown-dir ./Links \
  -db-path ./linky.db \
  -email shameer@example.com \
  -password mysecurepassword
```

The script will:
- Create `linky.db` in the current directory
- Read all `.md` files from `./Links`
- Convert filenames to Subjects (e.g., `now.md` → "now" subject)
- Convert ### headings to Topics
- Import all links under their respective topics

### Step 3: Start the Server

```bash
# Using environment variables (recommended for production)
JWT_SECRET="your-secret-key-here" \
DB_PATH="./linky.db" \
MARKDOWN_DIR="./Links" \
./bin/linky

# Or just run with defaults
./bin/linky
```

Environment variables:
- `JWT_SECRET`: Secret key for JWT tokens (default: test key, change in production!)
- `DB_PATH`: Path to SQLite database (default: `./linky.db`)
- `MARKDOWN_DIR`: Path to markdown files (default: `./Links`, for legacy support)

### Step 4: Access the Application

1. Open your browser to `http://localhost:8080`
2. You'll see the login page
3. Sign in with the credentials you used in the migration script
4. Start using Linky!

## Features

### Authentication
- **Sign Up**: Create a new account
- **Login**: Access your bookmarks
- **Logout**: Sign out securely
- **JWT Tokens**: Secure, stateless authentication

### Search Improvements
- **OR operator**: `python OR javascript` - matches links with either term
- **AND operator**: `tutorial AND beginner` or `tutorial & beginner` - matches links with both terms
- **Subject filter**: `subject:learn python` - search within a specific subject
- **Better scoring**: More relevant results at the top

### Keyboard Shortcuts
- `↑/↓` or `j/k`: Navigate links
- `x`: Toggle selection
- `Enter`: Open link
- `Delete`: Delete selected
- `Shift + C`: Copy selected
- `Shift + N`: Add bulk links
- `Shift + |`: Open selected links
- `?`: Show keyboard shortcuts
- `Ctrl + F`: Focus search
- `Ctrl + O`: Focus file selector

### API Endpoints

#### Public (No Authentication)
- `POST /api/auth/signup` - Create account
- `POST /api/auth/login` - Login

#### Legacy (Markdown Files)
- `GET /api/files` - List markdown files
- `GET /api/file/:filename` - Get links from file
- `POST /api/delete_links` - Delete links
- `POST /api/bulk_links` - Add bulk links

#### New API (SQLite, Requires Authentication)
All these endpoints require the `Authorization: Bearer <token>` header.

**Subjects:**
- `GET /api/v2/subjects` - List your subjects
- `POST /api/v2/subjects` - Create subject
- `GET /api/v2/subjects/:id` - Get subject
- `DELETE /api/v2/subjects/:id` - Delete subject

**Topics:**
- `GET /api/v2/subjects/:subjectId/topics` - List topics
- `POST /api/v2/subjects/:subjectId/topics` - Create topic
- `GET /api/v2/topics/:id` - Get topic
- `DELETE /api/v2/topics/:id` - Delete topic

**Links:**
- `GET /api/v2/topics/:topicId/links` - List links
- `POST /api/v2/topics/:topicId/links` - Create link
- `POST /api/v2/topics/:topicId/links/bulk` - Bulk create links
- `GET /api/v2/links/:id` - Get link
- `DELETE /api/v2/links/:id` - Delete link

**Tags:**
- `GET /api/v2/tags` - List your tags
- `POST /api/v2/tags` - Create tag
- `POST /api/v2/links/:linkId/tags/:tagId` - Add tag to link
- `DELETE /api/v2/links/:linkId/tags/:tagId` - Remove tag from link
- `GET /api/v2/links/:linkId/tags` - Get link's tags

## Security Notes

### Production Deployment

1. **Change JWT Secret:**
   ```bash
   JWT_SECRET="$(openssl rand -base64 32)" ./bin/linky
   ```

2. **Use HTTPS:** Deploy behind a reverse proxy (nginx, caddy) with SSL

3. **Secure Database:** Set proper file permissions on `linky.db`
   ```bash
   chmod 600 linky.db
   ```

4. **Strong Passwords:** Enforce strong password requirements

5. **Backup Database:**
   ```bash
   # Simple backup
   cp linky.db linky.db.backup

   # Automated backups
   sqlite3 linky.db ".backup linky_backup_$(date +%Y%m%d).db"
   ```

## Troubleshooting

### Migration Issues

**Problem:** "Failed to create user: UNIQUE constraint failed"
**Solution:** User with that email already exists. Use a different email or delete the database and start over.

**Problem:** "Failed to parse file"
**Solution:** Check that your markdown files follow the format:
```markdown
### Topic Name
- [Link Title](https://url.com)
- [Another Link](https://another-url.com)
```

### Login Issues

**Problem:** "Invalid token"
**Solution:** Clear your browser's localStorage and login again
```javascript
// In browser console:
localStorage.clear()
```

**Problem:** "Invalid email or password"
**Solution:** Double-check credentials or create a new account

### Database Issues

**Problem:** Database locked
**Solution:** Only one process can write to SQLite at a time. Make sure you don't have multiple server instances running.

## Data Schema

The SQLite database has the following structure:

```sql
users (id, email, password_hash, created_at, updated_at)
subjects (id, user_id, name, created_at, updated_at)
topics (id, subject_id, name, created_at, updated_at)
links (id, topic_id, title, url, created_at, updated_at)
tags (id, user_id, name)
link_tags (link_id, tag_id)
```

All data is isolated by `user_id` - users can only access their own data.

## FAQ

**Q: Can I still use markdown files?**
A: Yes! The legacy API endpoints still work. However, we recommend migrating to the new SQLite backend for better performance and features.

**Q: Can multiple users share the same database?**
A: Yes! Each user has their own isolated data. They sign up with different emails.

**Q: What happens to my old markdown files?**
A: They're preserved. The migration script only reads them, it doesn't modify or delete them.

**Q: Can I export my data?**
A: You can backup the SQLite database file. We may add export features in the future.

**Q: Is my password secure?**
A: Yes, passwords are hashed using bcrypt before storage. Plain text passwords are never stored.

## Support

For issues or questions:
1. Check the logs when running the server
2. Verify your environment variables are set correctly
3. Make sure the database file has proper permissions
4. Check that port 8080 is not already in use

## Next Steps

After migration, consider:
1. Setting up automated database backups
2. Deploying behind HTTPS
3. Creating additional user accounts for team members
4. Exploring the new tag feature for better organization
