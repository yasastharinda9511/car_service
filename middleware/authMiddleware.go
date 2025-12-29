package middleware

import (
	"car_service/logger"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

const (
	permissionsKey = "permissions"
	userIDKey      = "user_id"
)

type AuthMiddleware struct {
	client             *http.Client
	introspectEndPoint string
}

func NewAuthMiddleware(introspectEndPoint string) *AuthMiddleware {
	return &AuthMiddleware{
		&http.Client{Timeout: 10 * time.Second}, introspectEndPoint,
	}
}

func (authMiddleware *AuthMiddleware) Authorize(next http.Handler, permission string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"permission": permission,
		}).Debug("Authorizing request")

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.WithFields(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"ip":     r.RemoteAddr,
			}).Warn("Missing authorization header")
			writeError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.WithFields(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"ip":     r.RemoteAddr,
			}).Warn("Invalid authorization header format")
			writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]

		res, err := authMiddleware.introspect(&token)

		if err != nil || res.StatusCode != http.StatusOK {
			logger.WithFields(map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"ip":          r.RemoteAddr,
				"status_code": res.StatusCode,
				"error":       err,
			}).Warn("Token introspection failed")
			writeError(w, res.StatusCode, "Invalid authorization header")
			return
		}

		claims, err := authMiddleware.decodeAndVerifyToken(token)

		if err != nil {
			logger.WithFields(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"ip":     r.RemoteAddr,
				"error":  err.Error(),
			}).Warn("Failed to decode/verify token")
			writeError(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}

		if !authMiddleware.hasPermission(claims.Permissions, permission) {
			logger.WithFields(map[string]interface{}{
				"method":              r.Method,
				"path":                r.URL.Path,
				"ip":                  r.RemoteAddr,
				"required_permission": permission,
				"user_permissions":    claims.Permissions,
				"user_id":             claims.Subject,
			}).Warn("Permission denied")
			writeError(w, http.StatusUnauthorized, "No Valid Permission")
			return
		}

		logger.WithFields(map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_id":    claims.Subject,
			"permission": permission,
		}).Info("Authorization successful")

		ctx := r.Context()
		ctx = context.WithValue(ctx, permissionsKey, claims.Permissions)
		ctx = context.WithValue(ctx, userIDKey, claims.Subject)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (authMiddleware *AuthMiddleware) decodeAndVerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// Parse WITHOUT verification using jwt.ParseUnverified
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	return claims, nil
}

func (authMiddleware *AuthMiddleware) introspect(token *string) (*http.Response, error) {
	logger.WithField("endpoint", authMiddleware.introspectEndPoint).Debug("Calling token introspection endpoint")

	req, err := http.NewRequest("GET", authMiddleware.introspectEndPoint, nil)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"endpoint": authMiddleware.introspectEndPoint,
			"error":    err.Error(),
		}).Error("Failed to create introspection request")
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+*token)

	resp, err := authMiddleware.client.Do(req)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"endpoint": authMiddleware.introspectEndPoint,
			"error":    err.Error(),
		}).Error("Introspection request failed")
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"endpoint":    authMiddleware.introspectEndPoint,
		"status_code": resp.StatusCode,
	}).Debug("Introspection request completed")

	return resp, nil
}

func (authMiddleware *AuthMiddleware) hasPermission(allPermissions []string, permission string) bool {
	for _, perm := range allPermissions {
		if perm == permission {
			return true
		}
	}
	return false
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// GetPermissionsFromContext retrieves the permissions from the request context
func GetPermissionsFromContext(ctx context.Context) ([]string, bool) {
	permissions, ok := ctx.Value(permissionsKey).([]string)
	return permissions, ok
}
