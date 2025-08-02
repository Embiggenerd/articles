package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Embiggenerd/articles/core"
	"github.com/Embiggenerd/articles/stores"
	"github.com/go-chi/render"

	"github.com/joho/godotenv"
)

// package documents

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
		Email    string
		Password string
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
			http.Error(w, "Failed to read", http.StatusInternalServerError)
			return
		}

		id, err := userStore.Create(r.Context(), &core.User{Email: userReq.Email, Password: userReq.Password})
		if err != nil {
			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, UserCreateResponse{ID: id})
		render.Status(r, http.StatusOK)
	}
}

func HandleLoginUser(userStore core.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// data := new(bytes.Buffer)
		// _, err := io.Copy(data, r.Body)
		// if err != nil {
		// 	http.Error(w, "Failed to copy", http.StatusInternalServerError)
		// 	return
		// }

		var user core.User
		// err = binary.Read(data, binary.LittleEndian, &user)
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Failed to read", http.StatusInternalServerError)
			return
		}

		authenticated, foundUser, err := userStore.FindEmailAndAuthenticate(r.Context(), user.Email, user.Password)
		if err != nil || !authenticated {
			http.Error(w, "Failed to authenticate", http.StatusUnauthorized)
			return
		}

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

func Run(cfg Config) {
	mux := http.NewServeMux()
	var frontendHandler http.HandlerFunc = http.FileServer(http.Dir("../front/dist")).ServeHTTP
	if cfg.GOENV == "dev" {

		log.Println("Proxying requests to Vite dev server on port 9090")
		proxy, err := NewProxy("http://localhost:3001")
		if err != nil {
			panic(err)
		}
		frontendHandler = ProxyRequestHandler(proxy)
		// handle all requests to your server using the proxy
	}
	store := stores.GetStore()
	mux.HandleFunc("/", frontendHandler)
	// mux.HandleFunc("/api/", apiHandler)
	mux.HandleFunc("/api/document/get", HandleGetDocument(store.Documents))
	mux.HandleFunc("/api/document/post", HandleCreateDocument(store.Documents))

	mux.HandleFunc("/api/user/create", HandleCreateUser(store.Users))
	mux.HandleFunc("/api/user/login", HandleLoginUser(store.Users))

	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err.Error())
	}

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
	Run(cfg)
}
