package utils

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"strings"
)

func GetEnv(key string, fallback string) string {
	key = strings.TrimSpace(key)
	fallback = strings.TrimSpace(fallback)
	value := strings.TrimSpace(os.Getenv(key))
	if len(value) == 0 {
		log.Printf("Could to get provided env var: %s, using default value instead: %s", key, fallback)
		return fallback
	}
	return value
}

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "GinContextKey", c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GinContextFromContext(ctx context.Context) (*gin.Context, error) {
	ginContext := ctx.Value("GinContextKey")
	if ginContext == nil {
		err := fmt.Errorf("could not retrieve gin.Context")
		return nil, err
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		err := fmt.Errorf("gin.Context has wrong type")
		return nil, err
	}
	return gc, nil
}
