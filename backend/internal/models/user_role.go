package models

type UserRole struct {
	BaseModel
	Name         string `json:"name" gorm:"type:varchar(30);not null;index:idx_user_roles_name"`
	DisplayOrder int    `json:"display_order" gorm:"not null;index:idx_user_roles_display_order"`
	IsActive     bool   `json:"is_active" gorm:"type:boolean;not null;default:true"`
}
