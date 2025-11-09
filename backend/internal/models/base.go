package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate = validator.New()

type Model interface {
	BeforeCreate(*gorm.DB) error
	BeforeUpdate(*gorm.DB) error
}

type BaseModel struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Time      TimeProvider
}

// テスト用に時刻のフィールドを依存関係の注入
type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

func (r *RealTimeProvider) Now() time.Time {
	return time.Now()
}

var realTimeProvider TimeProvider = &RealTimeProvider{}

// フック関数
func (b *BaseModel) BeforeCreate(g *gorm.DB) error {
	err := b.baseValidate(g.Statement.Dest)
	if err != nil {
		return err
	}

	b.CreatedAt = b.Time.Now()
	b.UpdatedAt = b.Time.Now()
	return nil
}

func (b *BaseModel) BeforeUpdate(g *gorm.DB) error {
	err := b.baseValidate(g.Statement.Dest)
	if err != nil {
		return err
	}

	b.UpdatedAt = b.Time.Now()
	return nil
}

func (b *BaseModel) baseValidate(model interface{}) error {
	return validate.Struct(model)
}
