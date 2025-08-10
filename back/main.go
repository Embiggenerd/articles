package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Embiggenerd/articles/config"
	"github.com/Embiggenerd/articles/core"
	"github.com/Embiggenerd/articles/logger"
	"github.com/Embiggenerd/articles/stores"
	"github.com/go-chi/render"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("secret-key")

func createToken(secret string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(secret), 14)
	return string(bytes), err
}

func verifyToken(secret, token string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(token), []byte(secret))
	return err == nil
}

type (
	DocumentCreateResponse struct {
		ID string `json:"id"`
	}
	UserCreateResponse struct {
		ID string `json:"id"`
	}

	UserLoginResponse struct {
		ID       string `json:"id,omitempty"`
		Email    string `json:"email,omitempty"`
		Username string `json:"username,omitempty"`
	}
	UserCreateRequest struct {
		Email    string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
	}

	CookieValue struct {
		User   core.User
		Secret string
	}
)

func HandleCreateDocument(documentStore core.DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := new(bytes.Buffer)
		_, err := io.Copy(data, r.Body)
		if err != nil {
			http.Error(w, "Failed to copy", http.StatusInternalServerError)
			return
		}
		id, err := documentStore.Create(r.Context(), &core.Document{Data: *data})
		if err != nil {
			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, DocumentCreateResponse{ID: id})
		render.Status(r, http.StatusOK)
	}
}

func HandleCreateUser(userStore core.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userReq UserCreateRequest
		err := json.NewDecoder(r.Body).Decode(&userReq)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		id, err := userStore.Create(r.Context(), &core.User{Email: userReq.Email, Password: userReq.Password})
		if err != nil {
			if sqlite3Err, ok := err.(sqlite3.Error); ok {
				if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
					http.Error(w, "User already exists", http.StatusConflict)
					return
				}
			}
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, UserCreateResponse{ID: id})
		render.Status(r, http.StatusOK)
	}
}

func HandleLoginUser(userStore core.UserStore, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user core.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Failed to read", http.StatusInternalServerError)
			return
		}

		// Check password matches hashed value stored in DB
		authenticated, foundUser, err := userStore.FindEmailAndAuthenticate(r.Context(), user.Email, user.Password)
		if err != nil || !authenticated {
			http.Error(w, "Failed to authenticate", http.StatusUnauthorized)
			return
		}

		// Create a cookie token we can decrypt later
		token, err := createToken(cfg.Get(cfg.AuthSecret))
		if err != nil {
			http.Error(w, "Failed to write", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     "exampleCookie",
			Value:    token,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}

		// Use the http.SetCookie() function to send the cookie to the client.
		// Behind the scenes this adds a `Set-Cookie` header to the response
		// containing the necessary cookie data.
		http.SetCookie(w, &cookie)

		render.JSON(w, r, UserLoginResponse{
			ID:       foundUser.ID,
			Email:    foundUser.Email,
			Username: foundUser.UserName,
		})
		render.Status(r, http.StatusOK)
	}
}

func HandleGetDocument(documentStore core.DocumentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		document, err := documentStore.FindID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Write(document.Data.Bytes())
	}
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

func HandleProtected(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("exampleCookie")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "cookie not found", http.StatusBadRequest)
			default:
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return
		}

		verified := verifyToken(cfg.Get(cfg.AuthSecret), cookie.Value)
		if !verified {
			http.Error(w, "Failed to authenticate", http.StatusUnauthorized)
			return
		}

		w.Write([]byte(cookie.Value))
	}
}

func Run(cfg config.Config, logr logger.Logger) {
	mux := http.NewServeMux()
	var frontendHandler http.HandlerFunc = http.FileServer(http.Dir("../front/dist")).ServeHTTP
	if cfg.Get(cfg.GoEnv) == "dev" {

		log.Println("Proxying requests to Vite dev server on port 9090")
		proxy, err := NewProxy("http://localhost:3001")
		if err != nil {
			panic(err)
		}
		frontendHandler = ProxyRequestHandler(proxy)
		// handle all requests to your server using the proxy
	}
	store := stores.GetStore(cfg)
	mux.HandleFunc("/", frontendHandler)
	// mux.HandleFunc("/api/", apiHandler)
	mux.HandleFunc("/api/document/get", HandleGetDocument(store.Documents))
	mux.HandleFunc("/api/document/post", HandleCreateDocument(store.Documents))

	mux.HandleFunc("/api/user/create", HandleCreateUser(store.Users))
	mux.HandleFunc("/api/user/login", HandleLoginUser(store.Users, cfg))
	mux.HandleFunc("/api/protected", HandleProtected(cfg))

	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err.Error())
	}

	withMW := logr.LoggingMW(mux)

	if err := http.Serve(listener, withMW); err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLoggerService(context.Background(), &cfg)
	fmt.Println("hihihi", log)
	Run(cfg, log)
}
