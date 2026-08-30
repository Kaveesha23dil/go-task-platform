package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kaveesha23dil/go-task-platform/internal/handler"
	"github.com/Kaveesha23dil/go-task-platform/internal/model"
	"github.com/Kaveesha23dil/go-task-platform/internal/repository"
	"github.com/Kaveesha23dil/go-task-platform/internal/service"
	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := service.NewTaskService(repository.NewMemoryTaskRepository())
	handler.RegisterRoutes(r, handler.NewTaskHandler(s))
	return r
}

func request(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func createTask(t *testing.T, router http.Handler) model.Task {
	t.Helper()
	response := request(router, http.MethodPost, "/api/v1/tasks", `{"type":"email","payload":"Send welcome email","priority":2}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	var task model.Task
	if err := json.Unmarshal(response.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode task: %v", err)
	}
	return task
}

func TestHealth(t *testing.T) {
	response := request(newTestRouter(), http.MethodGet, "/health", "")
	if response.Code != http.StatusOK || response.Body.String() != `{"service":"goflow-api","status":"ok"}` {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestCreateTask(t *testing.T) {
	task := createTask(t, newTestRouter())
	if task.ID == "" || task.Status != model.TaskStatusQueued {
		t.Fatalf("expected ID and queued status, got %+v", task)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	cases := []struct{ name, body, code string }{
		{"missing type", `{"payload":"work","priority":2}`, "INVALID_TASK"},
		{"invalid priority", `{"type":"email","payload":"work","priority":6}`, "INVALID_PRIORITY"},
		{"malformed JSON", `{`, "INVALID_REQUEST"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := request(newTestRouter(), http.MethodPost, "/api/v1/tasks", tc.body)
			if response.Code != http.StatusBadRequest || !bytes.Contains(response.Body.Bytes(), []byte(tc.code)) {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestListAndGetTasks(t *testing.T) {
	router := newTestRouter()
	created := createTask(t, router)
	response := request(router, http.MethodGet, "/api/v1/tasks", "")
	var tasks []model.Task
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &tasks) != nil || len(tasks) != 1 {
		t.Fatalf("unexpected list response: %d %s", response.Code, response.Body.String())
	}
	response = request(router, http.MethodGet, "/api/v1/tasks/"+created.ID, "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestGetNonexistentTask(t *testing.T) {
	response := request(newTestRouter(), http.MethodGet, "/api/v1/tasks/missing", "")
	if response.Code != http.StatusNotFound || !bytes.Contains(response.Body.Bytes(), []byte("TASK_NOT_FOUND")) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestDeleteTask(t *testing.T) {
	router := newTestRouter()
	created := createTask(t, router)
	response := request(router, http.MethodDelete, "/api/v1/tasks/"+created.ID, "")
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
	response = request(router, http.MethodDelete, "/api/v1/tasks/"+created.ID, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}
