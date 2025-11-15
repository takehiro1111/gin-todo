package config

import "time"

const (
	Development = "development"
	Staging     = "staging"
	Production  = "production"
)

type DBConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func GetDBConfig(env string) *DBConfig {
	switch env {
	case Production:
		return &DBConfig{
			MaxIdleConns:    25,
			MaxOpenConns:    200,
			ConnMaxLifetime: time.Hour,
		}
	case Staging:
		return &DBConfig{
			MaxIdleConns:    15,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
		}
	default:
		return &DBConfig{
			MaxIdleConns:    5,
			MaxOpenConns:    50,
			ConnMaxLifetime: 30 * time.Minute,
		}
	}
}
