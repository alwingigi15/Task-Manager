package repository

import (
	"Task-Manager/config"
	"Task-Manager/models"
	"database/sql"
	"fmt"
)

type TaskRepoInterface interface {
	CreateTask(task *models.Task) (*models.Task, error) 
	GetTasksForUser(userID string, limit, offset int, statusFilter string) ([]models.Task, error)
	GetAllTasks(limit, offset int, statusFilter string) ([]models.Task, error)
	GetTaskByID(id string) (*models.Task, error)
	UpdateTask(task *models.Task) (*models.Task, error)
	DeleteTask(id string) error
	AutoCompleteTask(id string) error 
}

type TaskRepository struct {
}

func (r *TaskRepository) CreateTask(task *models.Task) (*models.Task, error) {
	db, err := config.Dbconnection()
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO tasks (id, title, description, status, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, title, description, status, user_id, created_at, updated_at
	`

	var createdTask models.Task
	err = tx.QueryRow(query, task.ID, task.Title, task.Description, task.Status, task.UserID).
		Scan(&createdTask.ID, &createdTask.Title, &createdTask.Description, &createdTask.Status, &createdTask.UserID, &createdTask.CreatedAt, &createdTask.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &createdTask, nil
}

func (r *TaskRepository) GetTasksForUser(userID string, limit, offset int, statusFilter string) ([]models.Task, error) {
	db, err := config.Dbconnection()
	if err != nil {
		return nil, err
	}

	query := `SELECT id, title, description, status, user_id, created_at, updated_at FROM tasks WHERE user_id = $1`
	params := []interface{}{userID}

	if statusFilter != "" {
		query += " AND status = $2"
		params = append(params, statusFilter)
	}

	query += " ORDER BY created_at DESC LIMIT $3 OFFSET $4"
	params = append(params, limit, offset)

	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) GetAllTasks(limit, offset int, statusFilter string) ([]models.Task, error) {
	db, err := config.Dbconnection()
	if err != nil {
		return nil, err
	}

	query := `SELECT id, title, description, status, user_id, created_at, updated_at FROM tasks`
	params := []interface{}{}

	if statusFilter != "" {
		query += " WHERE status = $1"
		params = append(params, statusFilter)
	}

	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	if statusFilter == "" {
		params = append(params, limit, offset)
	} else {
		params = append(params[0:1], limit, offset)
		params = params
	}

	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("query all tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) GetTaskByID(id string) (*models.Task, error) {
	db, err := config.Dbconnection()
	if err != nil {
		return nil, err
	}

	query := `SELECT id, title, description, status, user_id, created_at, updated_at FROM tasks WHERE id = $1`

	var t models.Task
	err = db.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	} else if err != nil {
		return nil, fmt.Errorf("query task: %w", err)
	}

	return &t, nil
}

func (r *TaskRepository) UpdateTask(task *models.Task) (*models.Task, error) {
	db, err := config.Dbconnection()
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE tasks SET title = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, title, description, status, user_id, created_at, updated_at
	`

	var updatedTask models.Task
	err = tx.QueryRow(query, task.Title, task.Description, task.Status, task.ID).
		Scan(&updatedTask.ID, &updatedTask.Title, &updatedTask.Description, &updatedTask.Status, &updatedTask.UserID, &updatedTask.CreatedAt, &updatedTask.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &updatedTask, nil
}

func (r *TaskRepository) DeleteTask(id string) error {
	db, err := config.Dbconnection()
	if err != nil {
		return err
	}

	query := `DELETE FROM tasks WHERE id = $1`
	_, err = db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (r *TaskRepository) AutoCompleteTask(id string) error {
	db, err := config.Dbconnection()
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Check current status
	var currentStatus models.TaskStatus
	err = tx.QueryRow(`SELECT status FROM tasks WHERE id = $1`, id).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return nil // Task deleted or not found, skip
	} else if err != nil {
		return fmt.Errorf("check status: %w", err)
	}

	if currentStatus == models.StatusCompleted {
		return nil // Already completed
	}

	// Update to completed
	_, err = tx.Exec(`UPDATE tasks SET status = $1, updated_at = NOW() WHERE id = $2`, models.StatusCompleted, id)
	if err != nil {
		return fmt.Errorf("auto-complete: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
