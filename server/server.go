package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
var dbClient *mongo.Database

func main() {
	port := getEnv("PORT", defaultPort)
	hostURL := fmt.Sprintf("%s:%s", getEnv("HOST_URL", defaultHostUrl), port)
	redisHost := getEnv("REDIS_HOST", defaultRedisHost)
	redisPass := getEnv("REDIS_PASS", defaultRedisPassword)
	redisNumConn, err := strconv.Atoi(getEnv("REDIS_NUM_CONN", defaultRedisNumConn))
	if err != nil {
		log.Fatalf("Could not parse number of idle redis connections: %s", err)
	}
	useSecureCookies, err := strconv.ParseBool(getEnv("USE_SECURE_COOKIES", "false"))
	if err!=nil{
		log.Fatalf("Could not parse useSecureCookies bool: %s", err)
	}

	// initialize db client
	dbClient = retrieveDatabaseConn()
	// Setting up Gin
	r := gin.Default()

	store, err := redis.NewStore(redisNumConn, "tcp", redisHost, redisPass,
		[]byte(getEnv("COOKIE_SECRET", string(securecookie.GenerateRandomKey(256)))))

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
func retrieveDatabaseConn() *mongo.Database {
	dbURL := getEnv("DB_URL", defaultDbUrl)
	dbName := getEnv("DB_NAME", defaultDbName)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURL))
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Database could not be pinged: %s", err)
	}
	return client.Database(dbName)
}
