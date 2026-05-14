package router

import (
	gql "github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func GraphQLRoutes(api fiber.Router, resolver *gql.Resolver) {
	server := gql.GraphQLServer(resolver)

	api.All("/graphql", adaptor.HTTPHandler(server))
}