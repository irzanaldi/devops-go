package repository

import (
	"github.com/irzan-aldi/task-manager/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) FindByID(id uint) (*model.Task, error) {
	var task model.Task
	err := r.db.Preload("Assignee").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) FindByBoardID(boardID uint, page, limit int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.Where("board_id = ?", boardID)
	query.Model(&model.Task{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Assignee").Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRepository) FindByAssigneeID(assigneeID uint, page, limit int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.Where("assignee_id = ?", assigneeID)
	query.Model(&model.Task{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Assignee").Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(id uint) error {
	return r.db.Delete(&model.Task{}, id).Error
}
