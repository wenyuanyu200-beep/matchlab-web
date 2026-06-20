// Package database creates PostgreSQL connections for the application.
package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"matchlab/backend/internal/config"
)

// Open creates and verifies a GORM PostgreSQL connection.
func Open(cfg config.Database) (*gorm.DB, error) {
	if !cfg.Configured() {
		return nil, fmt.Errorf("database configuration is incomplete")
	}

	db, err := gorm.Open(postgres.Open(DSN(cfg)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access PostgreSQL connection pool: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return db, nil
}

// DSN returns a libpq-style PostgreSQL data source name.
func DSN(cfg config.Database) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		quote(cfg.Host), quote(cfg.Port), quote(cfg.User), quote(cfg.Password),
		quote(cfg.Name), quote(cfg.SSLMode),
	)
}

// SafeDescription identifies a database connection without exposing secrets.
func SafeDescription(cfg config.Database) string {
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s", cfg.Host, cfg.Port, cfg.Name, cfg.User)
}

func quote(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	if strings.ContainsAny(escaped, " \t\n\r'") {
		return "'" + escaped + "'"
	}
	return escaped
}
