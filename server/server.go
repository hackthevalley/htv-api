package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		err := os.Setenv("PORT", defaultPort)
		if err != nil {
			log.Fatalf("Could not set default PORT: %s", err)
		}
	}

	// Setting up Gin
	r := gin.Default()
	privateRoutes := r.Group("/v1")
	privateRoutes.Use(authMiddleware())
	privateRoutes.POST("/graphql", graphqlHandler())
	privateRoutes.GET("/playground", playgroundHandler())

	publicRoutes := r.Group("/v1")
	publicRoutes.GET("/auth/callback", authCallback)

	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
	//http.HandleFunc("/auth/callback", authCallback)
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
