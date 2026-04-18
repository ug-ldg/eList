package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ug-ldg/elist/internal/repository"
)

type StatsHandler struct {
	repo *repository.StatsRepository
}

func NewStatsHandler(repo *repository.StatsRepository) *StatsHandler {
	return &StatsHandler{repo: repo}
}

func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetStats(r.Context())

	if err != nil {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
