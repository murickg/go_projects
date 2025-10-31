package service

import (
	"strings"

	"example.com/taskmanager/internal/model"
	"example.com/taskmanager/internal/repository"
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(r repository.TaskRepository) *TaskService {
	return &TaskService{repo: r}
}

func (s *TaskService) Create(title, desc string) (*model.Task, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrInvalidInput
	}
	t := &model.Task{
		Title:       title,
		Description: desc,
		Done:        false,
	}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Get(id int) (*model.Task, error) {
	t, err := s.repo.GetByID(id)
	if err == repository.ErrNotFound {
		return nil, ErrNotFound
	}
	return t, err
}

func (s *TaskService) List() ([]model.Task, error) {
	return s.repo.List()
}

func (s *TaskService) Update(id int, title, desc *string, done *bool) (*model.Task, error) {
	t, err := s.repo.GetByID(id)
	if err == repository.ErrNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	if title != nil && strings.TrimSpace(*title) != "" {
		t.Title = *title
	}
	if desc != nil {
		t.Description = *desc
	}
	if done != nil {
		t.Done = *done
	}
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Delete(id int) error {
	if err := s.repo.Delete(id); err == repository.ErrNotFound {
		return ErrNotFound
	} else {
		return err
	}
}
