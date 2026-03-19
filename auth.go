// auth.go

package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("secret")

type contextKey string

const (
	userIDContextKey   contextKey = "userID"
	userRoleContextKey contextKey = "userRole"
)

type AuthClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func tokenFromAuthorizationHeader(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}

	return value
}

// ValidateJWT validates a JWT token and extracts the user ID and role.
func ValidateJWT(tokenString string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims.Subject, claims.Role, nil
	}

	return "", "", jwt.ErrSignatureInvalid
}

func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDContextKey).(string)
	return userID
}

func UserRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(userRoleContextKey).(string)
	if role == "" {
		return "user"
	}
	return role
}

// Middleware to protect routes with JWT authentication
func WithJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := tokenFromAuthorizationHeader(r.Header.Get("Authorization"))
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		userID, role, err := ValidateJWT(tokenString)
		if err != nil || userID == "" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		ctx = context.WithValue(ctx, userRoleContextKey, role)

		next(w, r.WithContext(ctx))
	}
}
