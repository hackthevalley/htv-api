package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hackthevalley/htv-api/database"
	"github.com/hackthevalley/htv-api/utils"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
		clientID := strings.TrimSpace(utils.GetEnv("client_id", ""))
		redirectURL := strings.TrimSpace(utils.GetEnv("redirect_uri", ""))
		responseType := strings.TrimSpace(utils.GetEnv("response_type", ""))
		scope := strings.TrimSpace(utils.GetEnv("scope", ""))
		s := sessions.Default(c)
		authToken := s.Get("htv-token")
		log.Printf("Recieved auth token: %v", authToken)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userFilter := &bson.M{"sessionID": authToken}
		query := database.DbClient.Collection("users").FindOne(ctx, userFilter)
		log.Printf("Find query: %v", query.Err())
		if query.Err() != nil {
			log.Printf("No auth cookie present, redirecting user to login")
			u, err := url.Parse("https://my.mlh.io/oauth/authorize")
			if err != nil {
				log.Print(err)
			}
			u.RawQuery = fmt.Sprintf("client_id=%s&redirect_uri=%s&response_type=%s&scope=%s&state=%s",
				clientID,
				redirectURL,
				responseType,
				scope,
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
		q.Add("client_id", utils.GetEnv("client_id", ""))
		q.Add("client_secret", utils.GetEnv("client_secret", ""))
		q.Add("redirect_uri", utils.GetEnv("redirect_uri", ""))
		q.Add("grant_type", utils.GetEnv("grant_type", ""))
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
	sessionToken, err := uuid.NewRandom()
	sessionTokenStr := sessionToken.String()
	if err != nil {
		log.Printf("Could not generate random session token: %s", err)
		return
	}
	log.Printf("Generated Token: %s", sessionTokenStr)
	s.Set("htv-token", sessionTokenStr)
	err = s.Save()
	if err != nil {
		log.Printf("Failed to save session token: %s", err)
		return
	}
	checkToken := s.Get("htv-token")
	log.Printf("Check token: %v", checkToken)
	err = createUser(authToken, sessionTokenStr)
	if err != nil {
		log.Printf("Could not create user: %s", err)
		return
	}
	redirectUserState(c)
}
func createUser(authToken AuthToken, sessionToken string) error {
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
		profileMap, ok := gjson.Parse(gjson.Get(string(profileBody), "data").String()).Value().(map[string]interface{})
		if !ok {
			log.Printf("Failed to unmarshal profile json, during profile creation")
			return err
		}
		userFilter := &bson.M{"email": userEmail}
		query := database.DbClient.Collection("users").FindOne(ctx, userFilter)
		log.Printf("Find query: %v", query.Err())
		if query.Err() != nil {
			res, err := database.DbClient.Collection("users").InsertOne(ctx, bson.M{
				"idpProfile":   profileMap,
				"sessionID": sessionToken,
				"email":userEmail,
			})
			if err != nil {
				log.Printf("Could not insert user into database: %s", err)
				return nil
			}
			log.Printf("Inserted user to database: %v", res)
		} else {
			res, err := database.DbClient.Collection("users").ReplaceOne(ctx, userFilter, bson.M{
				"email":userEmail,
				"idpProfile":   profileMap,
				"sessionID": sessionToken,
			})
			if err != nil {
				log.Printf("Could not insert user into database: %s", err)
				return nil
			}
			log.Printf("Inserted user to database: %v", res)
		}
	}
	return err
}
