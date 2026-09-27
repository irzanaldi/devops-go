package repository

import (
	"github.com/irzan-aldi/task-manager/internal/model"
	"gorm.io/gorm"
)

type BoardRepository struct {
	db *gorm.DB
}

func NewBoardRepository(db *gorm.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

func (r *BoardRepository) Create(board *model.Board) error {
	return r.db.Create(board).Error
}

func (r *BoardRepository) FindByID(id uint) (*model.Board, error) {
	var board model.Board
	err := r.db.Preload("Owner").Preload("Tasks").First(&board, id).Error
	if err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *BoardRepository) FindByOwnerID(ownerID uint, page, limit int) ([]model.Board, int64, error) {
	var boards []model.Board
	var total int64

	query := r.db.Where("owner_id = ?", ownerID)
	query.Model(&model.Board{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Tasks").Offset(offset).Limit(limit).Find(&boards).Error
	return boards, total, err
}

func (r *BoardRepository) Update(board *model.Board) error {
	return r.db.Save(board).Error
}

func (r *BoardRepository) Delete(id uint) error {
	return r.db.Delete(&model.Board{}, id).Error
}
