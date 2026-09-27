package service

import (
	"errors"

	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/repository"
)

type BoardService struct {
	repo *repository.BoardRepository
}

func NewBoardService(repo *repository.BoardRepository) *BoardService {
	return &BoardService{repo: repo}
}

func (s *BoardService) Create(ownerID uint, req model.CreateBoardRequest) (*model.Board, error) {
	board := &model.Board{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(board); err != nil {
		return nil, errors.New("failed to create board")
	}
	return board, nil
}

func (s *BoardService) GetByID(id uint) (*model.Board, error) {
	board, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("board not found")
	}
	return board, nil
}

func (s *BoardService) GetByOwnerID(ownerID uint, page, limit int) ([]model.Board, int64, error) {
	return s.repo.FindByOwnerID(ownerID, page, limit)
}

func (s *BoardService) Update(id, ownerID uint, req model.UpdateBoardRequest) (*model.Board, error) {
	board, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("board not found")
	}

	if board.OwnerID != ownerID {
		return nil, errors.New("not authorized to update this board")
	}

	if req.Title != nil {
		board.Title = *req.Title
	}
	if req.Description != nil {
		board.Description = *req.Description
	}

	if err := s.repo.Update(board); err != nil {
		return nil, errors.New("failed to update board")
	}
	return board, nil
}

func (s *BoardService) Delete(id, ownerID uint) error {
	board, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("board not found")
	}

	if board.OwnerID != ownerID {
		return errors.New("not authorized to delete this board")
	}

	return s.repo.Delete(id)
}
