package repository

import "example.com/taskmanager/internal/model"

type TaskRepository interface {
	Create(t *model.Task) error
	GetByID(id int) (*model.Task, error)
	List() ([]model.Task, error)
	Update(t *model.Task) error
	Delete(id int) error
}
