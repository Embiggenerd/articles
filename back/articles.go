package main

import (
	"fmt"
	"log"
	"net/http"

	twitterLogin "github.com/dghubble/gologin/v2/twitter"
	"github.com/dghubble/oauth1"
)

// Define your Consumer Key and Consumer Secret
const (
	ConsumerKey    = "YOUR_CONSUMER_KEY"
	ConsumerSecret = "YOUR_CONSUMER_SECRET"
)

func main() {
	// OAuth1 Config for Twitter
	oauth1Config := &oauth1.Config{
		ConsumerKey:    ConsumerKey,
		ConsumerSecret: ConsumerSecret,
		CallbackURL:    "http://localhost:8080/twitter/callback",
		Endpoint:       oauth1.TwitterEndpoint,
	}

	// Success handler for authenticated users
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		twitterUser, err := twitterLogin.UserFromContext(ctx)
		if err != nil {
			http.Error(w, "Failed to get Twitter user from context", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Welcome, %s!", twitterUser.ScreenName)
	})

	// Failure handler for authentication errors
	failureHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Twitter login failed.")
	})

	// Setup routes
	http.Handle("/twitter/login", twitterLogin.LoginHandler(oauth1Config, nil))
	http.Handle("/twitter/callback", twitterLogin.CallbackHandler(oauth1Config, successHandler, failureHandler))

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
