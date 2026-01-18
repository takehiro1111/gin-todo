package models

import "time"

type PasswordResetToken struct {
	BaseModel
	ResetToken string    `json:"reset_token" gorm:"type:varchar(1000);not null;uniqueIndex"`
	UserID     uint      `gorm:"not null;index"`
	ExpiresAt  time.Time `gorm:"not null"`
	// nullを許容するためポインタを参照している
	UsedAt *time.Time `gorm:"default:null"`

	// Relation
	User *User `gorm:"foreignKey:UserID"`
}
