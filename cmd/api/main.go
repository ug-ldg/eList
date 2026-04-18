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

	r := chi.NewRouter()
	r.Post("/tasks", taskHandler.Create)
	r.Get("/tasks/{id}", taskHandler.Get)
	r.Get("/tasks/{id}/children", taskHandler.GetChildren)
	r.Patch("/tasks/{id}/status", taskHandler.UpdateStatus)

	r.Get("/stats", statsHandler.Get)

	port := os.Getenv("PORT")
	fmt.Printf("server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
