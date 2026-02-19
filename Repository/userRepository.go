package repository

import (
	"Task-Manager/config"
	"Task-Manager/models"
	"database/sql"
	"fmt"
)

type UserRepoInterface interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}

type UserRepository struct {
}

func (r *UserRepository) CreateUser(user *models.User) (*models.User, error) {
	db, err := config.Dbconnection()
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (id, email, username, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, email, username, role, created_at, updated_at
	`

	var createdUser models.User
	err = db.QueryRow(query, user.ID, user.Email, user.Username, user.Password, user.Role).
		Scan(&createdUser.ID, &createdUser.Email, &createdUser.Username, &createdUser.Role, &createdUser.CreatedAt, &createdUser.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &createdUser, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	db, err := config.Dbconnection()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, email, username, password, role, created_at, updated_at
		FROM users WHERE email = $1
	`

	var user models.User
	err = db.QueryRow(query, email).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}
