package main

import (
	"fmt"
	"log"
	"net/http"

	hhttp "example.com/taskmanager/internal/handler/http"
	"example.com/taskmanager/internal/repository"
	"example.com/taskmanager/internal/service"
)

func main() {
	// Wiring зависимостей
	repo := repository.NewTaskRepoMemory()
	tasks := service.NewTaskService(repo)
	auth := service.NewAuthService()
	h := hhttp.NewTaskHandler(tasks, auth)
	rtr := hhttp.NewRouter(h, auth)

	fmt.Println("Task Manager running on http://localhost:8080 (demo creds: admin/admin)")
	log.Fatal(http.ListenAndServe(":8080", rtr))
}
