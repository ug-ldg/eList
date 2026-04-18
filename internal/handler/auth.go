package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/ug-ldg/elist/internal/auth"
	"github.com/ug-ldg/elist/internal/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	userRepo    *repository.UserRepository
	googleOAuth *oauth2.Config
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	cfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &AuthHandler{userRepo: userRepo, googleOAuth: cfg}
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.googleOAuth.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := h.googleOAuth.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := h.googleOAuth.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var info struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		http.Error(w, "failed to decode user info", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.Upsert(r.Context(), "google", info.ID, info.Email, info.Name)
	if err != nil {
		http.Error(w, "failed to upsert user", http.StatusInternalServerError)
		return
	}

	jwt, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": jwt})
}
