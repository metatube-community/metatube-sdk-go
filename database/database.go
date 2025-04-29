package database

import (
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	Sqlite   = "sqlite"
	Postgres = "postgres"
)

type Config struct {
	// DSN the Data Source Name.
	DSN string

	// Disable automatic ping.
	DisableAutomaticPing bool

	// Prepared statement.
	PreparedStmt bool

	// Max DB open connections.
	MaxOpenConns int

	// Max DB idle connections.
	MaxIdleConns int

	LogLevel logger.LogLevel
}

func (cfg *Config) applyDefaults() {
	if cfg.DSN == "" {
		// use sqlite DB memory mode by default.
		cfg.DSN = "file::memory:?cache=shared"
	}

	if cfg.MaxIdleConns <= 0 {
		// golang's default.
		cfg.MaxIdleConns = 2
	}

	if cfg.LogLevel < logger.Silent ||
		cfg.LogLevel > logger.Info {
		// INFO by default.
		cfg.LogLevel = logger.Info
	}
}

func Open(cfg *Config) (*gorm.DB, error) {
	cfg.applyDefaults()

	var dialector gorm.Dialector
	// We try to parse it as postgresql, otherwise
	// fallback to sqlite.
	if regexp.MustCompile(`^postgres(ql)?://`).MatchString(cfg.DSN) ||
		len(strings.Fields(cfg.DSN)) >= 3 {
		dialector = postgres.New(postgres.Config{
			DSN: cfg.DSN,
			// set true to disable implicit prepared statement usage.
			PreferSimpleProtocol: !cfg.PreparedStmt,
		})
	} else {
		dialector = sqlite.Open(cfg.DSN)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "[GORM]\u0020", log.LstdFlags),
			logger.Config{
				SlowThreshold:             100 * time.Millisecond,
				LogLevel:                  cfg.LogLevel,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      false,
				Colorful:                  false,
			}),
		PrepareStmt:          cfg.PreparedStmt,
		DisableAutomaticPing: cfg.DisableAutomaticPing,
	})
	if err != nil {
		return nil, err
	}

	if sqlDB, err := db.DB(); err == nil /* ignore error */ {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	return db, nil
}
