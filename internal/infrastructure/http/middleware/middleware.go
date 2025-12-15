package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// เก็บ config
type Middleware struct {
	cfg *config.Config
}

// constructor
func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{cfg: cfg}
}

func (m *Middleware) Recover() fiber.Handler {
	return recover.New()
}

func (m *Middleware) RequestID() fiber.Handler {
	return requestid.New()
}

// --- CORS Middleware ---
func (m *Middleware) CORS() fiber.Handler {
	allowedOrigins := "*"
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	})
}

// --- Request Logging Middleware ---
func (m *Middleware) RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		rid, _ := c.Locals("requestid").(string)
		uid, _ := c.Locals("user_id").(string)

		status := c.Response().StatusCode()

		evt := log.Info()
		if status >= 500 {
			evt = log.Error()
		} else if status >= 400 {
			evt = log.Warn()
		}

		e := evt.
			Str("request_id", rid).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.IP())

		if strings.TrimSpace(uid) != "" {
			e = e.Str("user_id", uid)
		}

		e.Msg("HTTP Request")
		return err
	}
}

// --- Auth Middleware ---
func (m *Middleware) AuthRequired() fiber.Handler {
	secret := strings.TrimSpace(m.cfg.JWTSecret)
	if secret == "" {
		log.Fatal().Msg("JWT_SECRET is missing in config")
	}

	return func(c *fiber.Ctx) error {
		tokenStr, ok := bearerToken(c.Get("Authorization"))
		if !ok {
			return response.SendError(c, fiber.StatusUnauthorized, response.ErrUnauthorized, "Missing or invalid Authorization header")
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(
			tokenStr,
			claims,
			func(t *jwt.Token) (any, error) {
				if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)

		if err != nil || token == nil || !token.Valid {
			return response.SendError(c, fiber.StatusUnauthorized, response.ErrUnauthorized, "Invalid or expired token")
		}

		refID, ok := claimString(claims, "ref_id")
		if !ok || strings.TrimSpace(refID) == "" {
			return response.SendError(c, fiber.StatusUnauthorized, response.ErrUnauthorized, "Missing user_id claim")
		}

		role, _ := claimString(claims, "role")

		c.Locals("ref_id", refID)
		c.Locals("role", role)

		return c.Next()
	}
}

// returns the JWT token from "Authorization: Bearer <token>" (avoids repeated parsing)
func bearerToken(authHeader string) (string, bool) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}
	tok := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tok == "" {
		return "", false
	}
	return tok, true
}

// gets a claim as string safely (prevents panic if the claim isn't a string)
func claimString(claims jwt.MapClaims, key string) (string, bool) {
	v, ok := claims[key]
	if !ok || v == nil {
		return "", false
	}
	switch x := v.(type) {
	case string:
		return x, true
	default:
		s := strings.TrimSpace(fmt.Sprint(x))
		if s == "" || s == "<nil>" {
			return "", false
		}
		return s, true
	}
}
