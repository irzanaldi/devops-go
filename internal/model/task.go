package model

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Task struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      TaskStatus     `json:"status" gorm:"size:20;default:todo"`
	Priority    Priority       `json:"priority" gorm:"size:20;default:medium"`
	BoardID     uint           `json:"board_id" gorm:"not null"`
	Board       Board          `json:"board,omitempty" gorm:"foreignKey:BoardID"`
	AssigneeID  *uint          `json:"assignee_id"`
	Assignee    *User          `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
	DueDate     *time.Time     `json:"due_date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=200"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority" binding:"omitempty,oneof=low medium high"`
	AssigneeID  *uint      `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string     `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string     `json:"description"`
	Status      *TaskStatus `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority    *Priority   `json:"priority" binding:"omitempty,oneof=low medium high"`
	AssigneeID  *uint       `json:"assignee_id"`
	DueDate     *time.Time  `json:"due_date"`
}
