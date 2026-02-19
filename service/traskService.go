package service

import (
	repository "Task-Manager/Repository"
	//"Task-Manager/config"
	"Task-Manager/models"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type TaskServiceInterface interface {
	worker(delay int)
	ScheduleAutoComplete(id string)
	CreateTask(task *models.Task) (*models.Task, error) 
	GetTasks(userID, role string, limit, offset int, statusFilter string) ([]models.Task, error)
	GetTaskByID(id, userID, role string) (*models.Task, error)
	UpdateTask(task *models.Task, userID, role string) (*models.Task, error)
	DeleteTask(id, userID, role string) error 
}

type TaskService struct {
	autoCompleteCh chan string
}

func NewTaskService() *TaskService {
	delayMinutes := 5 // default

    delayStr := os.Getenv("AUTO_COMPLETE_MINUTES")
    if d, err := strconv.Atoi(delayStr); err == nil && d > 0 {
        delayMinutes = d
    }

    log.Printf("Auto-complete delay set to %d minutes", delayMinutes)

    ts := &TaskService{
        autoCompleteCh: make(chan string, 500),
    }

    go ts.worker(delayMinutes)

    return ts
}

func (s *TaskService) worker(delay int) {
	log.Println("Auto-complete worker started")
    Taskrepo := repository.TaskRepoInterface(&repository.TaskRepository{})
    for id := range s.autoCompleteCh {
        log.Printf("Worker received task %s for auto-complete", id)
        go func(taskID string) {
            log.Printf("Sleeping %d minutes for task %s", delay, taskID)
            time.Sleep(time.Duration(delay) * time.Minute)
            if err := Taskrepo.AutoCompleteTask(taskID); err != nil {
                log.Printf("Auto-complete failed for %s: %v", taskID, err)
            } else {
                log.Printf("Auto-completed task %s", taskID)
            }
        }(id)
    }
    log.Println("Auto-complete worker stopped")
}

func (s *TaskService) ScheduleAutoComplete(id string) {
	select {
    case s.autoCompleteCh <- id:
        log.Printf("Scheduled auto-complete for task %s", id)
    default:
        log.Printf("WARNING: auto-complete channel full - skipping task %s", id)
        // Optionally: send to a dead-letter queue or just ignore
    }
}

func (s *TaskService) CreateTask(task *models.Task) (*models.Task, error) {
	Taskrepo := repository.TaskRepoInterface(&repository.TaskRepository{})
	return Taskrepo.CreateTask(task)
}

func (s *TaskService) GetTasks(userID, role string, limit, offset int, statusFilter string) ([]models.Task, error) {
	Taskrepo := repository.TaskRepoInterface(&repository.TaskRepository{})
	if role == "admin" {
		return Taskrepo.GetAllTasks(limit, offset, statusFilter)
	}
	return Taskrepo.GetTasksForUser(userID, limit, offset, statusFilter)
}

func (s *TaskService) GetTaskByID(id, userID, role string) (*models.Task, error) {
	Taskrepo := repository.TaskRepoInterface(&repository.TaskRepository{})
	task, err := Taskrepo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	if role != "admin" && task.UserID != userID {
		return nil, fmt.Errorf("unauthorized to access this task")
	}

	return task, nil
}

func (s *TaskService) UpdateTask(task *models.Task, userID, role string) (*models.Task, error) {
	Taskrepo := repository.TaskRepoInterface(&repository.TaskRepository{})
	existing, err := Taskrepo.GetTaskByID(task.ID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && existing.UserID != userID {
		return nil, fmt.Errorf("unauthorized to update this task")
	}

	// Merge updates
	if task.Title != "" {
		existing.Title = task.Title
	}
	if task.Description != "" {
		existing.Description = task.Description
	}
	if task.Status != "" {
		existing.Status = task.Status
	}

	return Taskrepo.UpdateTask(existing)
}

func (s *TaskService) DeleteTask(id, userID, role string) error {
	Taskrepo := repository.TaskRepoInterface(&repository.TaskRepository{})
	existing, err := Taskrepo.GetTaskByID(id)
	if err != nil {
		return err
	}

	if role != "admin" && existing.UserID != userID {
		return fmt.Errorf("unauthorized to delete this task")
	}

	return Taskrepo.DeleteTask(id)
}
