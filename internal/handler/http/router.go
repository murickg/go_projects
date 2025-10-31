package http

import (
	"net/http"

	"example.com/taskmanager/internal/handler/middleware"
	"example.com/taskmanager/internal/service"
)

func NewRouter(h *TaskHandler, auth *service.AuthService) http.Handler {
	mux := http.NewServeMux()

	// Публичные
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)

	// Защищённые /tasks*
	tasks := http.NewServeMux()
	tasks.HandleFunc("GET /tasks", h.ListTasks)
	tasks.HandleFunc("POST /tasks", h.CreateTask)
	tasks.HandleFunc("GET /tasks/{id}", h.GetTask)
	tasks.HandleFunc("PUT /tasks/{id}", h.UpdateTask)
	tasks.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)

	// навешиваем middleware на поддерево /tasks
	smw := middleware.Session(auth)
	mux.Handle("/tasks", smw(tasks))
	mux.Handle("/tasks/", smw(tasks))

	var handler http.Handler = mux
	handler = middleware.Recovery(handler)
	handler = middleware.CORS(handler)
	handler = middleware.Logging(handler)
	return handler
}
