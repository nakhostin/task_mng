package entity

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Summary     string    `gorm:"not null"`
	Description string    `gorm:"not null"`
	Assignee    uint      `gorm:"not null"` // user id for foreign key
	Status      Status    `gorm:"not null"`
	Priority    Priority  `gorm:"not null"`
	DueDate     time.Time `gorm:"not null"`
}

func (Task) TableName() string {
	return "tasks"
}
