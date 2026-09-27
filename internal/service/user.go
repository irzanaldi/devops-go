package service

import (
	"errors"

	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(id uint) (*model.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	resp := user.ToResponse()
	return &resp, nil
}

func (s *UserService) GetAll(page, limit int) ([]model.UserResponse, int64, error) {
	users, total, err := s.repo.FindAll(page, limit)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.UserResponse
	for _, u := range users {
		responses = append(responses, u.ToResponse())
	}
	return responses, total, nil
}

func (s *UserService) Update(id uint, name string) (*model.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Name = name
	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("failed to update user")
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *UserService) Delete(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	return s.repo.Delete(id)
}
