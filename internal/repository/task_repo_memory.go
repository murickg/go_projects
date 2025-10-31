package repository

import (
	"errors"
	"sync"
	"time"

	"example.com/taskmanager/internal/model"
)

var ErrNotFound = errors.New("not found")

type TaskRepoMemory struct {
	mu   sync.RWMutex
	seq  int
	data map[int]model.Task
}

func NewTaskRepoMemory() *TaskRepoMemory {
	return &TaskRepoMemory{data: make(map[int]model.Task)}
}

func (r *TaskRepoMemory) Create(t *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	t.ID = r.seq
	t.CreatedAt = time.Now()
	r.data[t.ID] = *t
	return nil
}

func (r *TaskRepoMemory) GetByID(id int) (*model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	vv := v
	return &vv, nil
}

func (r *TaskRepoMemory) List() ([]model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Task, 0, len(r.data))
	for _, v := range r.data {
		out = append(out, v)
	}
	return out, nil
}

func (r *TaskRepoMemory) Update(t *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[t.ID]; !ok {
		return ErrNotFound
	}
	r.data[t.ID] = *t
	return nil
}

func (r *TaskRepoMemory) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return ErrNotFound
	}
	delete(r.data, id)
	return nil
}
