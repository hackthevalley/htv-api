package main

import (
	"io/ioutil"

	//"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.GetHeader("Authorization")
		if authToken != "" {
			log.Println(authToken)
			c.Next()
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Authorization Token Not Valid!"})
	}
}

func authCallback(c *gin.Context) {
	authorizationCode := c.Request.URL.Query().Get("code")
	if len(authorizationCode) > 0 {
		req, err := http.NewRequest("POST", "https://my.mlh.io/oauth/token", nil)
		if err != nil {
			log.Fatal(err)
		}
		q := req.URL.Query()
		q.Add("client_id", strings.TrimSpace(os.Getenv("client_id")))
		q.Add("client_secret", strings.TrimSpace(os.Getenv("client_secret")))
		q.Add("redirect_uri", strings.TrimSpace(os.Getenv("redirect_uri")))
		q.Add("grant_type", "authorization_code")
		q.Add("code", authorizationCode)
		req.URL.RawQuery = q.Encode()

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Request to MLH token endpoint failed: %s", err)
			return
		}

		authBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("Could not read in authCode: %s", err)
		}
		err = response.Body.Close()
		if err != nil {
			log.Printf("Could not close io read of response body: %s", err)
		}
		//err = json.Unmarshal(authPayload, &respToken)
		//if err!=nil{
		//	log.Println(err)
		//}
		log.Println(string(authBody))
		c.JSON(http.StatusOK, string(authBody))
	}
}
