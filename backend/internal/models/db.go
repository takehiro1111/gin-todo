package models

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DBInit(user, passWord, host, port, dbName, sslMode, tz, env string, intMaxIdleConns, intMaxOpenConns, intConnMaxLifetime int) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", host, user, passWord, dbName, port, sslMode, tz)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection failed:%w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(intMaxIdleConns)
	sqlDB.SetMaxOpenConns(intMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(intConnMaxLifetime) * time.Second)

	return db, nil
}

func RunMigration(user, password, host, port, dbName, sslMode string) error {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbName, sslMode)
	m, err := migrate.New(
		"file://migrations",
		dbUrl,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Println("Migration failed, rolling back...")
		// マイグレーションに失敗した際はテーブルを削除する
		if downErr := m.Down(); downErr != nil {
			return fmt.Errorf("migration failed and rollback failed: %w, %v", err, downErr)
		}
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}
