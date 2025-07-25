package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// twitterLogin "github.com/dghubble/gologin/v2/twitter"
// "github.com/dghubble/oauth1"

// Define your Consumer Key and Consumer Secret
// const (
// 	ConsumerKey    = "X_CONSUMER_KEY"
// 	ConsumerSecret = "X_CONSUMER_SECRET"
// )

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	person := Person{Name: "John", Age: 30}

	// Encoding - One step
	jsonStr, err := json.Marshal(person)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)
}

func serveStatic() http.Handler {
	// ctx := r.Context()
	// reqID, _ := utils.ExposeContextMetadata(ctx).Get("requestID")
	// if r.URL.Path != "/" {
	// 	// s.log.LogRequestError(reqID.(string), "Not found", http.StatusNotFound)
	// 	http.Error(w, "Not found", http.StatusNotFound)
	// 	return
	fmt.Println("we are serving static files")
	// }
	// if r.Method != http.MethodGet {
	// 	// s.log.LogRequestError(reqID.(string), "Method not allowed", http.StatusMethodNotAllowed)
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }
	// http.ServeFile(w, r, "../front/dist/index.html")
	return http.FileServer(http.Dir("../front/dist"))
}

func Run(cfg Config) {
	mux := http.NewServeMux()
	var frontendHandler http.HandlerFunc = http.FileServer(http.Dir("../front/dist")).ServeHTTP
	// fs := http.FileServer(http.Dir("../front"))
	if cfg.GOENV == "dev" {
		// viteURL, _ := url.Parse("http://localhost:3001") // Replace with your Vite dev server URL
		// proxy := httputil.NewSingleHostReverseProxy(viteURL)
		// http.Handle("/", proxy)

		// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 	proxy.ServeHTTP(w, r)
		// })

		log.Println("Proxying requests to Vite dev server on port 8080")
		// log.Fatal(http.ListenAndServe(":8080", nil))
		// mux.Handle("/", fs)
		proxy, err := NewProxy("http://localhost:3001")
		if err != nil {
			panic(err)
		}
		frontendHandler = ProxyRequestHandler(proxy)
		// handle all requests to your server using the proxy
	}
	mux.HandleFunc("/", frontendHandler)
	mux.HandleFunc("/api/", apiHandler)

	// mux.HandleFunc("/", serveHome)
	// mux.HandleFunc("/ws", s.serveWS)

	// withMW := s.log.LoggingMW(mux)

	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err.Error())
	}

	// log.Info("server listening on port " + s.server.Addr)

	if err := http.Serve(l, mux); err != nil {
		log.Fatal(err.Error())
	}
}

type Config struct {
	XConsumerKey    string
	XConsumerSecret string
	GOENV           string
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()

	cfg := Config{
		XConsumerKey:    os.Getenv("X_CONSUMER_KEY"),
		XConsumerSecret: os.Getenv("X_CONSUMER_SECRET"),
		GOENV:           os.Getenv("GOENV"),
	}

	return cfg, err
}

func main() {
	cfg, err := LoadConfig()
	fmt.Println(err)

	fmt.Printf("%+v\n", cfg)

	// OAuth1 Config for Twitter
	// oauth1Config := &oauth1.Config{
	// 	ConsumerKey:    ConsumerKey,
	// 	ConsumerSecret: ConsumerSecret,
	// 	CallbackURL:    "http://localhost:8080/twitter/callback",
	// 	Endpoint:       oauth1.TwitterEndpoint,
	// }

	// // Success handler for authenticated users
	// successHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 	ctx := req.Context()
	// 	twitterUser, err := twitterLogin.UserFromContext(ctx)
	// 	if err != nil {
	// 		http.Error(w, "Failed to get Twitter user from context", http.StatusInternalServerError)
	// 		return
	// 	}
	// 	fmt.Fprintf(w, "Welcome, %s!", twitterUser.ScreenName)
	// })

	// // Failure handler for authentication errors
	// failureHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 	fmt.Fprintln(w, "Twitter login failed.")
	// })

	// // Setup routes
	// http.Handle("/twitter/login", twitterLogin.LoginHandler(oauth1Config, nil))
	// http.Handle("/twitter/callback", twitterLogin.CallbackHandler(oauth1Config, successHandler, failureHandler))

	// fmt.Println("Server listening on :8080")
	// http.HandleFunc("/api", homeHandler)
	// http.HandleFunc("/", serveHome)

	Run(cfg)
}
