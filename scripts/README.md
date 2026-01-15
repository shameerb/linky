# Data Migration Scripts

Scripts to migrate data from SQLite to Supabase.

## Prerequisites

- Go (for export script)
- Node.js (for import script)
- Your Supabase project URL and Service Role Key

## Step 1: Export from SQLite

Export your data from the SQLite database:

```bash
# Make sure you're in the scripts directory
cd scripts

# Run the export script
go run export-sqlite-to-json.go
```

This will create `linky-export.json` with all your data.

## Step 2: Install Dependencies

```bash
npm install
```

## Step 3: Import to Supabase

Get your Supabase credentials:
1. Go to your Supabase project dashboard
2. Navigate to Settings > API
3. Copy the **Project URL** and **Service Role Key** (NOT the anon key!)

Run the import:

```bash
SUPABASE_URL=https://xxxxx.supabase.co \
SUPABASE_SERVICE_KEY=your-service-role-key \
node import-to-supabase.js
```

Replace:
- `https://xxxxx.supabase.co` with your actual Project URL
- `your-service-role-key` with your actual Service Role Key

## Important Notes

⚠️ **Users will need to reset their passwords** after migration since we can't migrate password hashes to Supabase Auth. The import script creates users with temporary passwords.

✅ All data relationships (subjects → topics → links) are preserved with new UUIDs.

## Verification

After import, verify in Supabase:
1. Go to Table Editor
2. Check that all tables have data
3. Verify the counts match the export summary
