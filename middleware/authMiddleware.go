package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

const permissionsKey = "permissions"

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
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]

		res, err := authMiddleware.introspect(&token)

		if err != nil || res.StatusCode != http.StatusOK {
			writeError(w, res.StatusCode, "Invalid authorization header")
			return
		}

		claims, err := authMiddleware.decodeAndVerifyToken(token)

		if err != nil {
			writeError(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}

		if !authMiddleware.hasPermission(claims.Permissions, permission) {
			writeError(w, http.StatusUnauthorized, "No Valid Permission")
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, permissionsKey, claims.Permissions)

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

	req, err := http.NewRequest("GET", authMiddleware.introspectEndPoint, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+*token)

	return authMiddleware.client.Do(req)

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
