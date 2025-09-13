package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/karan-bishtt/auth-service/internal/utils"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Extract token from header
		tokenString, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			http.Error(w, `{"status": 401, "message": "Unauthorized: Invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Validate token
		userID, role, err := utils.GetUserFromToken(tokenString)
		if err != nil {
			http.Error(w, `{"status": 401, "message": "Unauthorized: Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, role)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission middleware checks if user has specific permission
func RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(uint)
			if !ok {
				http.Error(w, `{"status": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Check permission in database (you'll implement this)
			hasPermission := checkUserPermission(userID, resource, action)
			if !hasPermission {
				http.Error(w, `{"status": 403, "message": "Forbidden: Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks if user has specific role
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				http.Error(w, `{"status": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Check if user role is in allowed roles
			roleAllowed := false
			for _, role := range allowedRoles {
				if strings.EqualFold(userRole, role) {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				http.Error(w, `{"status": 403, "message": "Forbidden: Insufficient role"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to check user permission (implement based on your needs)
func checkUserPermission(userID uint, resource, action string) bool {
	// This would query your database to check if user has the specific permission
	// For now, returning true - you'll implement the actual logic
	return true
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	return userID, ok
}

// GetUserRoleFromContext extracts user role from request context
func GetUserRoleFromContext(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(UserRoleKey).(string)
	return role, ok
}
