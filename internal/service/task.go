package service

import (
	"errors"

	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/repository"
)

type TaskService struct {
	repo      *repository.TaskRepository
	boardRepo *repository.BoardRepository
}

func NewTaskService(repo *repository.TaskRepository, boardRepo *repository.BoardRepository) *TaskService {
	return &TaskService{repo: repo, boardRepo: boardRepo}
}

func (s *TaskService) Create(boardID uint, userID uint, req model.CreateTaskRequest) (*model.Task, error) {
	// Verify board exists and user has access
	board, err := s.boardRepo.FindByID(boardID)
	if err != nil {
		return nil, errors.New("board not found")
	}

	if board.OwnerID != userID {
		return nil, errors.New("not authorized to add tasks to this board")
	}

	task := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		BoardID:     boardID,
		AssigneeID:  req.AssigneeID,
		DueDate:     req.DueDate,
		Status:      model.StatusTodo,
	}

	if task.Priority == "" {
		task.Priority = model.PriorityMedium
	}

	if err := s.repo.Create(task); err != nil {
		return nil, errors.New("failed to create task")
	}
	return task, nil
}

func (s *TaskService) GetByID(id uint) (*model.Task, error) {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (s *TaskService) GetByBoardID(boardID uint, page, limit int) ([]model.Task, int64, error) {
	return s.repo.FindByBoardID(boardID, page, limit)
}

func (s *TaskService) GetMyTasks(userID uint, page, limit int) ([]model.Task, int64, error) {
	return s.repo.FindByAssigneeID(userID, page, limit)
}

func (s *TaskService) Update(id, userID uint, req model.UpdateTaskRequest) (*model.Task, error) {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	// Check board ownership
	board, err := s.boardRepo.FindByID(task.BoardID)
	if err != nil {
		return nil, errors.New("board not found")
	}

	// Allow board owner or task assignee to update
	if board.OwnerID != userID && (task.AssigneeID == nil || *task.AssigneeID != userID) {
		return nil, errors.New("not authorized to update this task")
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := s.repo.Update(task); err != nil {
		return nil, errors.New("failed to update task")
	}
	return task, nil
}

func (s *TaskService) Delete(id, userID uint) error {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("task not found")
	}

	board, err := s.boardRepo.FindByID(task.BoardID)
	if err != nil {
		return errors.New("board not found")
	}

	if board.OwnerID != userID {
		return errors.New("not authorized to delete this task")
	}

	return s.repo.Delete(id)
}
