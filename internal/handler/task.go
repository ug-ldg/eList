package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ug-ldg/elist/internal/middleware"
	"github.com/ug-ldg/elist/internal/service"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string `json:"title"`
		ParentID *int   `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)
	task, err := h.svc.CreateTask(r.Context(), userID, input.Title, input.ParentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)
	task, fromCache, err := h.svc.GetTask(r.Context(), userID, id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	if fromCache {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)
	task, err := h.svc.GetChildren(r.Context(), userID, id)
	if err != nil {
		http.Error(w, "failed to get children", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var input struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)
	task, err := h.svc.UpdateTaskStatus(r.Context(), userID, id, input.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)
	err = h.svc.DeleteTask(r.Context(), userID, id)
	if err != nil {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)
	task, err := h.svc.GetTree(r.Context(), userID, id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetRootTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	task, err := h.svc.GetRootTasks(r.Context(), userID)
	if err != nil {
		http.Error(w, "root tasks not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
