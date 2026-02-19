package middleware

import (
	"Task-Manager/config"
	"Task-Manager/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/time/rate"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		log.Println("Auth header received:", authHeader)
		if authHeader == "" {
			utils.HandleError(w, nil, "Missing Authorization header", http.StatusUnauthorized)
			return // ← VERY IMPORTANT: STOP HERE
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader { // no "Bearer " prefix
			utils.HandleError(w, nil, "Invalid Authorization format. Use Bearer <token>", http.StatusUnauthorized)
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			utils.HandleError(w, err, "Server configuration error", http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			utils.HandleError(w, err, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.HandleError(w, nil, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			utils.HandleError(w, nil, "user_id missing in token", http.StatusUnauthorized)
			return
		}

		role, _ := claims["role"].(string) // role can be empty for safety

		// Now set context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "role", role)

		// ONLY if everything is ok → call next
		log.Println("Auth successful, calling next handler")
		next(w, r.WithContext(ctx))
	}
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	limitStr := os.Getenv("RATE_LIMIT")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10 // Default 10 reqs/min
	}

	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(limit)), limit)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			utils.HandleError(w, nil, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
