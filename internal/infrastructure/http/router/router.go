package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	gql "github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler/graphql"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handler.AllOfHandler, mw *middleware.Middleware, gqlResolver *gql.Resolver) {
	api := app.Group("/api")

	AuthRoutes(api, h, mw)
	EventRoutes(api, h, mw)
	HealthCheckRoutes(api, h)
	AuthRoutes(api, h, mw)
	EventRoutes(api, h, mw)
	GraphQLRoutes(api, gqlResolver)
}
