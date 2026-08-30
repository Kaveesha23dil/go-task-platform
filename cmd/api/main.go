package main

import (
	"github.com/Kaveesha23dil/go-task-platform/internal/handler"
	"github.com/Kaveesha23dil/go-task-platform/internal/repository"
	"github.com/Kaveesha23dil/go-task-platform/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	taskRepository := repository.NewMemoryTaskRepository()
	taskService := service.NewTaskService(taskRepository)
	taskHandler := handler.NewTaskHandler(taskService)

	handler.RegisterRoutes(router, taskHandler)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
