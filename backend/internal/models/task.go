package models

import (
	"time"
)

type Task struct {
	BaseModel
	UserID      uint       `json:"user_id" gorm:"index:idx_tasks_user_id;not null"`
	Title       string     `json:"title" gorm:"type:varchar(200);not null"`
	Description string     `json:"description" gorm:"type:text"`
	StatusID    int64      `json:"status_id" gorm:"index:idx_tasks_status_id;not null"`
	Priority    string     `json:"priority" gorm:"not null;default:medium"`
	DueDate     *time.Time `json:"due_date" gorm:"index:idx_tasks_due_date"` // SQL側でNULL制約をしているためポインタで指定

	// リレーション
	User   *User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	Status *TaskStatus `gorm:"foreignKey:StatusID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"status,omitempty"`
}
