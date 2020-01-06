package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/hackthevalley/htv-api/database"
	"github.com/hackthevalley/htv-api/utils"
	"log"
	"strconv"
	"time"
)

const defaultPort = "8080"
const defaultDbUrl = "mongodb://admin:password@localhost:27017"
const defaultDbName = "htv"
const defaultRedisHost = "localhost:6379"
const defaultRedisPassword = ""
const defaultRedisNumConn = "20"
const defaultHostUrl = "http://localhost"

func main() {
	port := utils.GetEnv("PORT", defaultPort)
	hostURL := fmt.Sprintf("%s:%s", utils.GetEnv("HOST_URL", defaultHostUrl), port)
	redisHost := utils.GetEnv("REDIS_HOST", defaultRedisHost)
	redisPass := utils.GetEnv("REDIS_PASS", defaultRedisPassword)
	redisNumConn, err := strconv.Atoi(utils.GetEnv("REDIS_NUM_CONN", defaultRedisNumConn))
	if err != nil {
		log.Fatalf("Could not parse number of idle redis connections: %s", err)
	}
	useSecureCookies, err := strconv.ParseBool(utils.GetEnv("USE_SECURE_COOKIES", "false"))
	if err != nil {
		log.Fatalf("Could not parse useSecureCookies bool: %s", err)
	}

	// initialize db client
	database.DbClient = database.RetrieveDatabaseConn(defaultDbUrl, defaultDbName)
	// Setting up Gin
	r := gin.Default()
	store, err := redis.NewStore(redisNumConn, "tcp", redisHost, redisPass,
		[]byte(utils.GetEnv("COOKIE_SECRET", string(securecookie.GenerateRandomKey(256)))))

	store.Options(sessions.Options{
		// expire session after 1 hour even though mymlh oauth token is valid for 2 hours
		MaxAge:   3600,
		Path:     "/",
		Secure:   useSecureCookies,
		HttpOnly: true,})

	r.Use(sessions.Sessions("session_store", store))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{hostURL},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(utils.GinContextToContextMiddleware())

	privateRoutes := r.Group("/v1")
	privateRoutes.Use(authMiddleware())
	privateRoutes.POST("/graphql", graphqlHandler())
	privateRoutes.GET("/playground", playgroundHandler())

	publicRoutes := r.Group("/v1")
	publicRoutes.GET("/auth/callback", authCallback)

	log.Printf("connect to %s/v1/graphql for querying GraphQL directly", hostURL)
	log.Printf("connect to %s/v1/playground for GraphQL playground", hostURL)

	err = r.Run()
	if err != nil {
		log.Fatalf("Could not run Gin router: %s", err)
	}
}
