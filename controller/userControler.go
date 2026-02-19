package controller

import (
	"Task-Manager/config"
	"Task-Manager/models"
	"Task-Manager/utils"
	repository "Task-Manager/Repository"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
}

func (h *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.HandleError(w, err, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if user.Email == "" || user.Username == "" || user.Password == "" {
		utils.HandleError(w, nil, "Missing required fields", http.StatusBadRequest)
		return
	}

	user.ID = uuid.New().String()
	user.Role = "user" // Default role

	// Hash password
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.HandleError(w, err, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPw)

	userRepo:=repository.UserRepoInterface(&repository.UserRepository{})
	createdUser, err := userRepo.CreateUser(&user)
	if err != nil {
		utils.HandleError(w, err, "Failed to register user", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, "User registered successfully", createdUser, http.StatusCreated)
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.HandleError(w, err, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	userRepo:=repository.UserRepoInterface(&repository.UserRepository{})
	user, err :=userRepo.GetUserByEmail(creds.Email)
	if err != nil {
		utils.HandleError(w, err, "User not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		utils.HandleError(w, err, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	cfg, _ := config.LoadConfig() // Assume no error for simplicity
	tokenStr, err := token.SignedString([]byte(cfg.JwtSecret))
	if err != nil {
		utils.HandleError(w, err, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, "Login successful", map[string]string{"token": tokenStr}, http.StatusOK)
}
