package models

type User struct {
	BaseModel
	Name         string `json:"name" gorm:"type:varchar(100);not null"`
	Email        string `json:"email" gorm:"type:varchar(100);not null;uniqueIndex"`
	PasswordHash string `json:"-" gorm:"type:varchar(500);not null"`
	RoleName     string `json:"role_name" gorm:"type:varchar(30);not null;default:'user'"`

	// Relation
	Tasks []Task    `gorm:"foreignKey:UserID" json:"tasks,omitempty"`
	Role  *UserRole `gorm:"foreignKey:RoleName;references:Name" json:"role,omitempty"`
}
