package database

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

//go:embed schema.sql
var schemaSQL string

//go:embed seed.sql
var seedSQL string

func Open(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Exec(schemaSQL).Error; err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := ensureReferenceData(db); err != nil {
		return nil, err
	}
	if err := ensureBootstrapData(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureReferenceData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			`INSERT INTO categories (id, name, display_order) VALUES (0, '未知', 0)
ON CONFLICT(id) DO UPDATE SET name = excluded.name, display_order = excluded.display_order`,
			`INSERT INTO categories (id, name, display_order) VALUES (1, '單行本', 1)
ON CONFLICT(id) DO UPDATE SET name = excluded.name, display_order = excluded.display_order`,
			`INSERT INTO categories (id, name, display_order) VALUES (2, '同人誌', 2)
ON CONFLICT(id) DO UPDATE SET name = excluded.name, display_order = excluded.display_order`,
			`INSERT INTO categories (id, name, display_order) VALUES (3, '雜誌短篇/彩頁', 3)
ON CONFLICT(id) DO UPDATE SET name = excluded.name, display_order = excluded.display_order`,
			`INSERT INTO categories (id, name, display_order) VALUES (4, 'CG', 4)
ON CONFLICT(id) DO UPDATE SET name = excluded.name, display_order = excluded.display_order`,
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func ensureBootstrapData(db *gorm.DB) error {
	var total int64
	if err := db.Raw(`SELECT COUNT(*) FROM comics`).Scan(&total).Error; err != nil {
		return fmt.Errorf("count comics: %w", err)
	}
	if total > 0 {
		return nil
	}
	if err := db.Exec(seedSQL).Error; err != nil {
		return fmt.Errorf("seed demo data: %w", err)
	}
	return nil
}
