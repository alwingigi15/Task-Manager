package controller

import (
	"Task-Manager/models"
	"Task-Manager/service"
	"Task-Manager/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TaskHandler struct {
}

var GlobalTaskService service.TaskServiceInterface

func (h *TaskHandler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateTaskHandler reached - step 1: start")

	userIDVal := r.Context().Value("user_id")
	log.Printf("step 2: user_id from context = %v", userIDVal)

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		log.Println("step 3: user_id invalid or missing")
		utils.HandleError(w, nil, "user_id missing in context", http.StatusUnauthorized)
		return
	}
	log.Println("step 4: user_id valid →", userID)

	var task models.Task
	log.Println("step 5: about to decode body")
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("step 5 ERROR: decode failed →", err)
		utils.HandleError(w, err, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	log.Println("step 6: body decoded successfully → title =", task.Title)

	if task.Title == "" {
		log.Println("step 7: title missing")
		utils.HandleError(w, nil, "Title is required", http.StatusBadRequest)
		return
	}
	log.Println("step 8: title is present")

	task.ID = uuid.New().String()
	task.UserID = userID
	task.Status = models.StatusPending
	log.Println("step 9: task prepared → ID =", task.ID)

	taskService := GlobalTaskService
	log.Println("step 10: calling CreateTask in repo")

	createdTask, err := taskService.CreateTask(&task)
	if err != nil {
		log.Println("step 10 ERROR: CreateTask failed →", err)
		utils.HandleError(w, err, "Failed to create task", http.StatusInternalServerError)
		return
	}
	log.Println("step 11: task created successfully")

	// taskService := service.TaskServiceInterface(&service.TaskService{})
	log.Println("step 12: scheduling auto-complete")
	taskService.ScheduleAutoComplete(task.ID)
	log.Println("step 13: auto-complete scheduled")

	log.Println("step 14: sending success response")
	utils.SuccessResponse(w, "Task created successfully", createdTask, http.StatusCreated)
	log.Println("step 15: response sent")
}

func (h *TaskHandler) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	log.Printf("step 2: user_id from context = %v", userIDVal)

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		log.Println("step 3: user_id invalid or missing")
		utils.HandleError(w, nil, "user_id missing in context", http.StatusUnauthorized)
		return
	}
	roleVal := r.Context().Value("role")
	role := "user" // default
	if roleVal != nil {
		if r, ok := roleVal.(string); ok {
			role = r
		}
	}

	// Parse pagination & filter
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	statusFilter := r.URL.Query().Get("status")

	limit := 10
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	taskService := GlobalTaskService
	tasks, err := taskService.GetTasks(userID, role, limit, offset, statusFilter)
	if err != nil {
		utils.HandleError(w, err, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, "Tasks fetched successfully", tasks, http.StatusOK)
}

func (h *TaskHandler) GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Context().Value("user_id").(string)
	role := r.Context().Value("role").(string)

	taskService := GlobalTaskService
	task, err := taskService.GetTaskByID(id, userID, role)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, err, "Task not found", http.StatusNotFound)
		} else {
			utils.HandleError(w, err, "Failed to fetch task", http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, "Task fetched successfully", task, http.StatusOK)
}

func (h *TaskHandler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userIDVal := r.Context().Value("user_id")
	log.Printf("step 2: user_id from context = %v", userIDVal)

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		log.Println("step 3: user_id invalid or missing")
		utils.HandleError(w, nil, "user_id missing in context", http.StatusUnauthorized)
		return
	}
	roleVal := r.Context().Value("role")
	role := "user" // default
	if roleVal != nil {
		if r, ok := roleVal.(string); ok {
			role = r
		}
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.HandleError(w, err, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	task.ID = id
	taskService := GlobalTaskService
	updatedTask, err := taskService.UpdateTask(&task, userID, role)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, err, "Task not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "unauthorized") {
			utils.HandleError(w, err, "Unauthorized to update this task", http.StatusForbidden)
		} else {
			utils.HandleError(w, err, "Failed to update task", http.StatusInternalServerError)
		}
		return
	}

	// If status changed to in_progress or still pending, reschedule
	if updatedTask.Status == models.StatusPending || updatedTask.Status == models.StatusInProgress {
		taskService.ScheduleAutoComplete(updatedTask.ID)
	}

	utils.SuccessResponse(w, "Task updated successfully", updatedTask, http.StatusOK)
}

func (h *TaskHandler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Context().Value("user_id").(string)
	role := r.Context().Value("role").(string)

	taskService := GlobalTaskService
	err := taskService.DeleteTask(id, userID, role)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.HandleError(w, err, "Task not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "unauthorized") {
			utils.HandleError(w, err, "Unauthorized to delete this task", http.StatusForbidden)
		} else {
			utils.HandleError(w, err, "Failed to delete task", http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, "Task deleted successfully", nil, http.StatusOK)
}
