package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//clientID := strings.TrimSpace(os.Getenv("client_id"))
		//redirectURI := strings.TrimSpace(os.Getenv("redirect_uri"))
		//scope := strings.TrimSpace(os.Getenv("scope"))
		htvOauthSessionToken, err := c.Cookie("htvOauthSessionToken")
		if err != nil {
			log.Printf("htvOauthSessionToken - %s", err.Error())
			//c.Redirect(http.StatusTemporaryRedirect,
			//	fmt.Sprintf("https://my.mlh.io/oauth/authorize?"+
			//		"client_id=%s&redirect_uri=%s&response_type=code&scope=%s", clientID, redirectURI, scope))
		}
		log.Println(htvOauthSessionToken)
		c.Next()
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
		authPayload, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("Could not read in authCode: %s", err)
		}
		err = response.Body.Close()
		if err != nil {
			log.Printf("Could not close io read of response body: %s", err)
		}
		log.Println(string(authPayload))
		//c.SetCookie("htvOauthSessionToken",
		//	string(authPayload),
		//	60*60*24,
		//	"/",
		//	"127.0.0.1",
		//	true, true)
		log.Println("Set htvOauthSessionToken cookie")
	}
}
