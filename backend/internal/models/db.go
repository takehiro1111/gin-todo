package models

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DBInit(user, passWord, host, port, dbName, sslMode, tz string) (*gorm.DB, error) {
	// ロガーのMiddleware実装次第で追加
	// 階層で共通のパッケージを使用したいのでgormのものは使用しない

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", host, user, passWord, dbName, port, sslMode, tz)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection failed:%w", err)
	}

	// デバッグモードの処理の追加も検討

	return db, nil
}
