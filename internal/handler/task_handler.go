package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kaveesha23dil/go-task-platform/internal/repository"
	"github.com/Kaveesha23dil/go-task-platform/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
}

type createTaskRequest struct {
	Type     string `json:"type"`
	Payload  string `json:"payload"`
	Priority int    `json:"priority"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{service: taskService}
}

func RegisterRoutes(router *gin.Engine, taskHandler *TaskHandler) {
	router.GET("/health", Health)
	tasks := router.Group("/api/v1/tasks")
	tasks.POST("", taskHandler.Create)
	tasks.GET("", taskHandler.GetAll)
	tasks.GET("/:id", taskHandler.GetByID)
	tasks.DELETE("/:id", taskHandler.Delete)
}

func (h *TaskHandler) Create(c *gin.Context) {
	var request createTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "Request body must contain valid JSON")
		return
	}

	task, err := h.service.Create(c.Request.Context(), service.CreateTaskInput{
		Type: request.Type, Payload: request.Payload, Priority: request.Priority,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTask):
			writeError(c, http.StatusBadRequest, "INVALID_TASK", err.Error())
		case errors.Is(err, service.ErrInvalidPriority):
			writeError(c, http.StatusBadRequest, "INVALID_PRIORITY", err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		}
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetAll(c *gin.Context) {
	tasks, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetByID(c *gin.Context) {
	task, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if errors.Is(err, repository.ErrTaskNotFound) {
		writeError(c, http.StatusNotFound, "TASK_NOT_FOUND", "Task was not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, repository.ErrTaskNotFound) {
		writeError(c, http.StatusNotFound, "TASK_NOT_FOUND", "Task was not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		return
	}
	c.Status(http.StatusNoContent)
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorBody{Error: errorDetail{Code: code, Message: message}})
}
