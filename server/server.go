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
		if err!=nil{
			log.Fatalf("Could not set default PORT: %s",err)
		}
	}

	// Setting up Gin
	r := gin.Default()
	r.Use(authMiddleware())
	v1 := r.Group("/v1")
	{
		v1.POST("/graphql", graphqlHandler())
		v1.GET("/playground", playgroundHandler())
		v1.GET("/auth/callback", authCallback)
	}
	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
	//http.HandleFunc("/auth/callback", authCallback)
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
