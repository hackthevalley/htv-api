package main

import (
	"github.com/99designs/gqlgen/handler"
	"github.com/gin-gonic/gin"
)

// Defining the Playground handler
func playgroundHandler() gin.HandlerFunc {
	h := handler.Playground("GraphQL playground", "/v1/graphql")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
