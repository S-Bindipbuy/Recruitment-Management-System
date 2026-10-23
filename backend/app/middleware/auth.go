package middleware

import (
	"ems/app/httperr"
	"ems/config"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64    `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, email string, roles []string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.AppConfig.JWTExpiration) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if path == "/api/auth/login" || path == "/api/auth/register" || strings.HasPrefix(path, "/api/components/") {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return unauthorized(c, "Missing authorization header")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return unauthorized(c, "Invalid authorization format")
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return unauthorized(c, "Invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_roles", claims.Roles)

		return c.Next()
	}
}

// unauthorized returns 401. For full-page navigations to server-rendered pages
// (no bearer token) it serves the app shell, which boots the client-side login.
// In all cases a non-nil error is returned so the remaining handler chain is
// skipped.
func unauthorized(c *fiber.Ctx, msg string) error {
	c.Set("Cache-Control", "no-store")
	if strings.HasPrefix(c.Path(), "/pages/") && c.Get("Authorization") == "" {
		if err := c.SendFile(config.AppConfig.FrontendPath + "/static/index.html"); err != nil {
			return httperr.New(fiber.StatusInternalServerError, "failed to serve app shell")
		}
		return httperr.Served(fiber.StatusUnauthorized, msg)
	}
	return httperr.New(fiber.StatusUnauthorized, msg)
}
