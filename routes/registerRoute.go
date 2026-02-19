package routes

import (
	"Task-Manager/controller"
	"Task-Manager/utils/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func MasterRoutes(router *mux.Router) {
	masterRoutes := router.PathPrefix("/master").Subrouter()

	registerUserRoutes(masterRoutes)
	registerTaskRoutes(masterRoutes)

}

func registerUserRoutes(routes *mux.Router) {
	UserRoutes := routes.PathPrefix("/user").Subrouter()
	UserHandler := &controller.UserHandler{}

	UserRoutes.Handle("/register", http.HandlerFunc(UserHandler.RegisterHandler)).Methods("POST")
	UserRoutes.Handle("/login", http.HandlerFunc(UserHandler.LoginHandler)).Methods("POST")

}

func registerTaskRoutes(routes *mux.Router) {
	TaskRoutes := routes.PathPrefix("/task").Subrouter()
	TaskHandler := &controller.TaskHandler{}

	TaskRoutes.Use(middleware.RateLimitMiddleware)
	// TaskRoutes.Use(middleware.AuthMiddleware)

	TaskRoutes.Handle("/add", middleware.AuthMiddleware(http.HandlerFunc(TaskHandler.CreateTaskHandler))).Methods("POST")
	TaskRoutes.Handle("/get", middleware.AuthMiddleware(http.HandlerFunc(TaskHandler.GetTasksHandler))).Methods("GET")
	TaskRoutes.Handle("/get/{id}", middleware.AuthMiddleware(http.HandlerFunc(TaskHandler.GetTaskByIDHandler))).Methods("GET")
	TaskRoutes.Handle("/update/{id}", middleware.AuthMiddleware(http.HandlerFunc(TaskHandler.UpdateTaskHandler))).Methods("PUT")
	TaskRoutes.Handle("/delete/{id}", middleware.AuthMiddleware(http.HandlerFunc(TaskHandler.DeleteTaskHandler))).Methods("DELETE")

}
