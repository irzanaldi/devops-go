package model

import (
	"time"

	"gorm.io/gorm"
)

type Board struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"size:500"`
	OwnerID     uint           `json:"owner_id" gorm:"not null"`
	Owner       User           `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Tasks       []Task         `json:"tasks,omitempty" gorm:"foreignKey:BoardID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateBoardRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateBoardRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}
