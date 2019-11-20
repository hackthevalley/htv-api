package main

import (
	"github.com/99designs/gqlgen/handler"
	"github.com/gin-gonic/gin"
	htv_api "github.com/hackthevalley/htv-api"
	"log"
)

// Defining the Graphql handler
func graphqlHandler() gin.HandlerFunc {
	h := handler.GraphQL(htv_api.NewExecutableSchema(htv_api.Config{Resolvers: &htv_api.Resolver{}}))
	log.Println("entered graphql handler")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
