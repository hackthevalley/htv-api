package main

import (
	"fmt"
	"github.com/99designs/gqlgen/handler"
	"github.com/hackthevalley/htv-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.Handle("/playground", authMiddleware(handler.Playground("GraphQL playground", "/query")))
	http.Handle("/query", authMiddleware(handler.GraphQL(htv_api.NewExecutableSchema(htv_api.Config{Resolvers: &htv_api.Resolver{}}))))
	http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		authorizationCode := r.URL.Query().Get("code")
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
				log.Println("Request to MLH token endpoint failed: ", err)
				return
			}
			authCode, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Println(err)
			}
			response.Body.Close()
			log.Println(string(authCode))
		}
	})
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}

func authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := strings.TrimSpace(os.Getenv("client_id"))
		//clientSecret := strings.TrimSpace(os.Getenv("client_secret"))
		redirectURI := strings.TrimSpace(os.Getenv("redirect_uri"))
		scope := strings.TrimSpace(os.Getenv("scope"))
		htvOauthSessionToken, err := r.Cookie("htvOauthSessionToken")
		if err != nil {
			log.Printf("htvOauthSessionToken - %s", err.Error())
			http.Redirect(w, r, fmt.Sprintf("https://my.mlh.io/oauth/authorize?"+
				"client_id=%s&redirect_uri=%s&response_type=code&scope=%s", clientID, redirectURI, scope), 303)
		}
		log.Println(htvOauthSessionToken)
		h.ServeHTTP(w, r)
	})
}
