package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ug-ldg/elist/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
