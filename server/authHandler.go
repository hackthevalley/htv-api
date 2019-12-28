package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type AuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	CreatedAt   int    `json:"created_at"`
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		authToken := s.Get("htv-token")
		log.Printf("Recieved auth token: %v", authToken)
		if authToken == nil {
			log.Printf("No auth cookie present, redirecting user to login")
			u, err := url.Parse("https://my.mlh.io/oauth/authorize")
			if err != nil {
				log.Print(err)
			}
			u.RawQuery = fmt.Sprintf("client_id=%s&redirect_uri=%s&response_type=%s&scope=%s&state=%s",
				strings.TrimSpace(os.Getenv("client_id")),
				strings.TrimSpace(os.Getenv("redirect_uri")),
				strings.TrimSpace(os.Getenv("response_type")),
				strings.TrimSpace(os.Getenv("scope")),
				c.Request.URL.Path)
			log.Println(u.String())
			c.Redirect(http.StatusTemporaryRedirect, u.String())
		}
		c.Next()
	}
}
func retrieveAuthCode(code string) ([] byte, error) {
	code = strings.TrimSpace(code)
	if len(code) > 0 {
		req, err := http.NewRequest("POST", "https://my.mlh.io/oauth/token", nil)
		if err != nil {
			log.Fatalf("Could not form request to MLH token endpoint: %s", err)
			return []byte(""), err
		}
		q := req.URL.Query()
		q.Add("client_id", getEnv("client_id", ""))
		q.Add("client_secret", getEnv("client_secret", ""))
		q.Add("redirect_uri", getEnv("redirect_uri", ""))
		q.Add("grant_type", getEnv("grant_type", ""))
		q.Add("code", code)
		req.URL.RawQuery = q.Encode()

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Request to MLH token endpoint failed: %s", err)
			return []byte(""), err
		}

		authBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("Could not read in auth code: %s", err)
			_ = response.Body.Close()
			return []byte(""), err
		}
		err = response.Body.Close()
		if err != nil {
			log.Printf("Could not close io read of auth response body: %s", err)
			return []byte(""), err
		}
		return authBody, err
	}
	return []byte(""), errors.New("no auth code provided")
}
func redirectUserState(c *gin.Context) {
	state := strings.TrimSpace(c.Request.URL.Query().Get("state"))
	if len(state) > 0 {
		log.Printf("Redirecting user back to: %s", state)
		c.Redirect(http.StatusTemporaryRedirect, state)
	}
	c.Redirect(http.StatusTemporaryRedirect, "/v1/playground")
}
func authCallback(c *gin.Context) {
	log.Printf(c.Request.URL.Query().Encode())
	s := sessions.Default(c)
	s.Clear()
	authBody, err := retrieveAuthCode(c.Request.URL.Query().Get("code"))
	authToken := AuthToken{}
	err = json.Unmarshal(authBody, &authToken)
	if err != nil {
		log.Printf("Could not unmarshal auth body to auth map: %s", err)
		return
	}
	log.Printf("%+v\n", authToken)
	sessionToken, err := mapUserSession(authToken)
	if err != nil {
		log.Printf("Could not map session token in database: %s", err)
	}
	s.Set("htv-token", sessionToken)
	err = s.Save()
	if err != nil {
		log.Printf("Failed to save session token: %s", err)
		return
	}
	checkToken := s.Get("htv-token")
	log.Printf("Check token: %v", checkToken)
	err = createUser(authToken)
	if err != nil {
		log.Printf("Could not create user: %s", err)
		return
	}
	redirectUserState(c)
}
func createUser(authToken AuthToken) error {
	req, err := http.NewRequest("GET", "https://my.mlh.io/api/v2/user.json", nil)
	if err != nil {
		log.Fatalf("Could not form request to MLH user profile endpoint: %s", err)
		return err
	}
	q := req.URL.Query()
	q.Add("access_token", authToken.AccessToken)
	req.URL.RawQuery = q.Encode()
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Request to MLH token endpoint failed: %s", err)
		return err
	}
	profileBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Could not read in auth code: %s", err)
		_ = response.Body.Close()
		return err
	}
	err = response.Body.Close()
	if err != nil {
		log.Printf("Could not close io read of auth response body: %s", err)
		return err
	}
	profileStatus := gjson.Get(string(profileBody), "status").String()
	if profileStatus == "OK" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userEmail := gjson.Get(string(profileBody), "data.email").String()
		res, err := dbClient.Collection("users").InsertOne(ctx, bson.M{
			"_id":  userEmail,
			"data": string(profileBody),
		})
		if err != nil {
			log.Printf("Could not insert user into database: %s", err)
			log.Printf("User probably already exists in database: %s", userEmail)
			// very very hacky
			// TODO: properly find existing users before inserting
			return nil
		}
		log.Printf("Inserted user to database: %v", res.InsertedID)
	}
	return err
}
func mapUserSession(authToken AuthToken) (string, error) {
	sessionToken, err := uuid.NewRandom()
	sessionTokenStr := sessionToken.String()
	if err != nil {
		log.Printf("Could not generate random session token: %s", err)
		return "", err
	}
	log.Printf("Generated Token: %s", sessionTokenStr)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := dbClient.Collection("user_tokens").InsertOne(ctx, bson.M{
		"_id": sessionTokenStr, "authToken": authToken,
	})
	if err != nil {
		log.Printf("Could not insert user token to database: %s", err)
		return "", err
	}
	log.Printf("Inserted user token document to database: %v", res.InsertedID)
	return sessionTokenStr, err
}
