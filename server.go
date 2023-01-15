package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/saki-engineering/graphql-sample/graph"
	"github.com/saki-engineering/graphql-sample/graph/services"
	"github.com/saki-engineering/graphql-sample/internal"
	"github.com/saki-engineering/graphql-sample/middlewares/auth"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/mattn/go-sqlite3"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	defaultPort = "8080"
	dbFile      = "./mygraphql.db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=on", dbFile))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	service := services.New(db)
	srv := handler.NewDefaultServer(internal.NewExecutableSchema(internal.Config{
		Resolvers: &graph.Resolver{
			Srv:     service,
			Loaders: graph.NewLoaders(service),
		},
		Directives: graph.Directive,
		Complexity: graph.ComplexityConfig(),
	}))
	srv.Use(extension.FixedComplexityLimit(10))
	// srv.AroundRootFields(func(ctx context.Context, next graphql.RootResolver) graphql.Marshaler {
	// 	log.Println("before RootResolver")
	// 	res := next(ctx)
	// 	defer func() {
	// 		var b bytes.Buffer
	// 		res.MarshalGQL(&b)
	// 		log.Println("after RootResolver", b.String())
	// 	}()
	// 	return res
	// })
	// srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	// 	log.Println("before OperationHandler")
	// 	res := next(ctx)
	// 	defer log.Println("after OperationHandler", res)
	// 	return res
	// })
	// srv.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	// 	log.Println("before ResponseHandler")
	// 	res := next(ctx)
	// 	defer log.Println("after ResponseHandler", res)
	// 	return res
	// })
	// srv.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	// 	log.Println("before Resolver")
	// 	res, err = next(ctx)
	// 	defer log.Println("after Resolver", res)
	// 	return
	// })

	boil.DebugMode = true

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", auth.AuthMiddleware(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
