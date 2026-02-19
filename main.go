package main

import (
	"Task-Manager/config"
	"Task-Manager/controller"
	"Task-Manager/routes"
	"Task-Manager/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/urfave/negroni"
	//_ "Task-Manager/docs"
)

func main() {
	fmt.Println("Task Manager!")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("can't load the Config:", err)
	}
	if cfg.Port == "" {
		cfg.Port = ":80"
	} else if cfg.Port[0] != ':' {
		cfg.Port = ":" + cfg.Port
	}

	config.InitDB(config.ReadDbConnection())
	defer config.CloseDB()

	if err := config.MigrationTable(); err != nil {
		log.Fatal("Error Performing migration:", err)
	}

	taskSvc := service.NewTaskService()
	controller.GlobalTaskService = taskSvc
	log.Println("Task auto-complete worker initialized")

	router := mux.NewRouter()
	routes.MasterRoutes(router)

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("Starting to walk through registered routes...")
	errr := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, errr := route.GetPathTemplate()
		if errr != nil {
			return errr
		}
		fmt.Println("Registered route:", pathTemplate)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through routes: %v\n", errr)
	}

	myCors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Authorization"},
	})

	n := negroni.Classic()
	n.Use(myCors)
	n.UseHandler(router)

	srv := http.Server{
		Addr:    cfg.Port,
		Handler: n,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
