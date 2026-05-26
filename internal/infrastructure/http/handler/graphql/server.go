package graphql

import (
	"fmt"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

func GraphQLServer(resolver *Resolver) *handler.Server {
	if resolver == nil {
		panic(fmt.Errorf("graphql resolver is nil"))
	}

	if resolver.Service == nil {
		panic(fmt.Errorf("graphql resolver service is nil"))
	}

	if resolver.Service.Dashboard == nil {
		panic(fmt.Errorf("dashboard service is nil"))
	}

	srv := handler.New(NewExecutableSchema(Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	return srv
}
