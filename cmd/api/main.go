package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/ug-ldg/elist/internal/cache"
	"github.com/ug-ldg/elist/internal/handler"
	appMiddleware "github.com/ug-ldg/elist/internal/middleware"
	"github.com/ug-ldg/elist/internal/repository"
	"github.com/ug-ldg/elist/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file")
	}

	pool, err := repository.NewPool()
	if err != nil {
		log.Fatal("failed to connect to PostgreSQL: ", err)
	}
	defer pool.Close()

	taskRepo := repository.NewTaskRepository(pool)
	taskCache := cache.NewTaskCache(os.Getenv("REDIS_ADDR"))
	taskSvc := service.NewTaskService(taskRepo, taskCache)
	taskHandler := handler.NewTaskHandler(taskSvc)
	statsRepo := repository.NewStatsRepository(pool)
	statsHandler := handler.NewStatsHandler(statsRepo)
	userRepo := repository.NewUserRepository(pool)
	authHandler := handler.NewAuthHandler(userRepo)

	r := chi.NewRouter()

	// Auth routes — public
	r.Get("/auth/google", authHandler.GoogleLogin)
	r.Get("/auth/google/callback", authHandler.GoogleCallback)

	// Task + stats routes — protected
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.Authenticate)

		r.Post("/tasks", taskHandler.Create)
		r.Get("/tasks/{id}", taskHandler.Get)
		r.Get("/tasks/{id}/children", taskHandler.GetChildren)
		r.Patch("/tasks/{id}/status", taskHandler.UpdateStatus)
		r.Delete("/tasks/{id}", taskHandler.DeleteTask)
		r.Get("/tasks/{id}/tree", taskHandler.GetTree)
		r.Get("/tasks", taskHandler.GetRootTasks)
		r.Get("/stats", statsHandler.Get)
	})

	port := os.Getenv("PORT")
	fmt.Printf("server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
