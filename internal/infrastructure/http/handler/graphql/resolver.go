package graphql

import (
	"github.com/cunex-club/quickattend-backend/internal/service"
)

type Resolver struct {
	Service *service.AllOfService
}