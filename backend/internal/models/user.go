package models

type User struct {
	BaseModel
	Name         string `json:"name" gorm:"type:varchar(100);not null"`
	Email        string `json:"email" gorm:"type:varchar(100);not null;uniqueIndex"`
	PasswordHash string `json:"-" gorm:"type:varchar(500);not null"`
	RoleID       int64  `json:"role_id" gorm:"not null;default:1"`

	// Relation
	Tasks []Task    `gorm:"foreignKey:UserID" json:"tasks,omitempty"`
	Role  *UserRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}
