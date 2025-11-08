package models

type TaskStatus struct {
	BaseModel
	Name         string `json:"name" gorm:"type:varchar(30);not null;uniqueIndex:idx_task_statuses_name"`
	Description  string `json:"description" gorm:"type:varchar(50);not null"`
	DisplayOrder int    `json:"display_order" gorm:"not null;index:idx_task_statuses_display_order"`
	IsActive     bool   `json:"is_active" gorm:"type:boolean;not null;default:true"`
}
