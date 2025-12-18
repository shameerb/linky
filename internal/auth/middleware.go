package auth

import (
	"context"
	"net/http"
	"os"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const EmailKey contextKey = "email"

// Default user for local development
const DefaultUserID = 1
const DefaultEmail = "shameer789@gmail.com"

// AuthMiddleware validates the JWT token and adds user info to the request context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token
		claims, err := ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the user ID from the request context
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}

// GetEmail extracts the email from the request context
func GetEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(EmailKey).(string)
	return email, ok
}

// NoAuthMiddleware bypasses authentication and uses a default user
// Useful for local development
func NoAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add default user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, DefaultUserID)
		ctx = context.WithValue(ctx, EmailKey, DefaultEmail)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAuthMiddleware returns the appropriate auth middleware based on DISABLE_AUTH env var
func GetAuthMiddleware() func(http.Handler) http.Handler {
	if os.Getenv("DISABLE_AUTH") == "true" {
		return NoAuthMiddleware
	}
	return AuthMiddleware
}
