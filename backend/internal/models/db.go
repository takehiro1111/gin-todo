package models

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DBInit(user, passWord, host, port, dbName, sslMode, tz string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", host, user, passWord, dbName, port, sslMode, tz)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection failed:%w", err)
	}

	err = runMigration(user, passWord, host, port, dbName, sslMode)
	// マイグレーションで変更のない場合はエラーにしない(migrate.ErrNoChange)
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func runMigration(user, password, host, port, dbName, sslMode string) error {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbName, sslMode)
	m, err := migrate.New(
		"file://migrations",
		dbUrl,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}
