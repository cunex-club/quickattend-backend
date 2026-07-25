package router

import (
	gql "github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler/graphql"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func GraphQLRoutes(api fiber.Router, resolver *gql.Resolver, mw *middleware.Middleware) {
	server := gql.GraphQLServer(resolver)

	api.All("/graphql", mw.AuthRequired(), adaptor.HTTPHandler(server))
}
