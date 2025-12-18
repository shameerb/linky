package store

import (
	"fmt"
	"os"

	"markdown-editor/internal/services/store/markdown"
	"markdown-editor/internal/services/store/markdownstore"
	"markdown-editor/internal/services/store/sqlite"
	// Import postgres when ready: "markdown-editor/internal/services/store/postgres"
)

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeFile     StorageType = "file"
	StorageTypeSQLite   StorageType = "sqlite"
	StorageTypePostgres StorageType = "postgres"
)

// Config holds the configuration for store initialization
type Config struct {
	Type        StorageType
	MarkdownDir string // For legacy markdown storage
	FileDataDir string // For file-based JSON storage
	DBPath      string // For SQLite
	DBURL       string // For PostgreSQL (e.g., "postgres://user:pass@localhost/dbname")
}

// NewStoreFromEnv creates a store based on environment variables
func NewStoreFromEnv() (Store, UserStore, error) {
	config := Config{
		Type:        StorageType(getEnv("STORAGE_TYPE", "file")),
		MarkdownDir: getEnv("MARKDOWN_DIR", "./Links"),
		FileDataDir: getEnv("FILE_DATA_DIR", "./data"),
		DBPath:      getEnv("DB_PATH", "./linky.db"),
		DBURL:       getEnv("DATABASE_URL", ""),
	}

	return NewStore(config)
}

// NewStore creates a store based on the provided configuration
func NewStore(config Config) (Store, UserStore, error) {
	switch config.Type {
	case StorageTypeFile:
		return newFileStore(config)
	case StorageTypeSQLite:
		return newSQLiteStore(config)
	case StorageTypePostgres:
		return newPostgresStore(config)
	default:
		return nil, nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}

// newFileStore initializes file-based storage
func newFileStore(config Config) (Store, UserStore, error) {
	// Use legacy markdown store for backward compatibility
	mdStore, err := markdown.New(config.MarkdownDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize markdown store: %w", err)
	}

	// Use markdown-based UserStore that reads/writes actual .md files
	mdUserStore, err := markdownstore.New(config.MarkdownDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize markdown user store: %w", err)
	}

	return mdStore, mdUserStore, nil
}

// newSQLiteStore initializes SQLite storage
func newSQLiteStore(config Config) (Store, UserStore, error) {
	sqliteStore, err := sqlite.New(config.DBPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize SQLite store: %w", err)
	}

	// SQLite store can also implement legacy Store interface if needed
	// For now, we return nil for legacy store
	return nil, sqliteStore, nil
}

// newPostgresStore initializes PostgreSQL storage
func newPostgresStore(config Config) (Store, UserStore, error) {
	// TODO: Implement PostgreSQL store
	// pgStore, err := postgres.New(config.DBURL)
	// if err != nil {
	//     return nil, nil, fmt.Errorf("failed to initialize Postgres store: %w", err)
	// }
	// return nil, pgStore, nil

	return nil, nil, fmt.Errorf("PostgreSQL storage not yet implemented")
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
