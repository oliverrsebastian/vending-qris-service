package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Dsn                      string `koanf:"dsn"`
	MaxIdleConns             int    `koanf:"max_idle_conns"`
	MaxOpenConns             int    `koanf:"max_open_conns"`
	MaxConnLifetimeInMinutes int    `koanf:"max_conn_lifetime_in_minutes"`
}

func Open(cfg DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Warn),
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	nativeDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	nativeDB.SetMaxIdleConns(cfg.MaxIdleConns)
	nativeDB.SetMaxOpenConns(cfg.MaxOpenConns)
	nativeDB.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifetimeInMinutes) * time.Minute)

	return db, nil
}
