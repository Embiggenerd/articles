package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Embiggenerd/articles/config"
	"github.com/Embiggenerd/articles/logger"
	"github.com/Embiggenerd/articles/stores"
)

type APIServer struct {
	server *http.Server
	// articleService articles.ArticleService
	logger logger.Logger
	store  stores.Store
	cfg    *config.Config
}

func NewServer(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	store stores.Store,
) *APIServer {
	server := &http.Server{
		Addr:              ":" + cfg.Get(cfg.Port),
		ReadHeaderTimeout: 3 * time.Second,
	}
	return &APIServer{
		server: server,
		logger: log,
		store:  store,
		cfg:    cfg,
	}
}

func (s *APIServer) Run() {
	mux := http.NewServeMux()
	var frontendHandler http.HandlerFunc = http.FileServer(http.Dir("../front/dist")).ServeHTTP
	if s.cfg.Get(s.cfg.GoEnv) == "dev" {

		log.Println("Proxying requests to Vite dev server on port 9090")
		proxy, err := newProxy("http://localhost:3001")
		if err != nil {
			panic(err)
		}
		frontendHandler = proxyRequestHandler(proxy)
		// handle all requests to your server using the proxy
	}
	mux.HandleFunc("/", frontendHandler)
	// mux.HandleFunc("/api/", apiHandler)
	mux.HandleFunc("/api/document/get", s.handleGetDocument())
	mux.HandleFunc("/api/document/post", s.handleCreateDocument())

	mux.HandleFunc("/api/user/create", s.handleCreateUser())
	mux.HandleFunc("/api/user/login", s.handleLoginUser())
	mux.HandleFunc("/api/protected", s.handleProtected())

	listener, err := net.Listen("tcp", ":"+s.cfg.Get(s.cfg.Port))
	if err != nil {
		log.Fatal(err.Error())
	}

	withMW := s.logger.LoggingMW(mux)

	if err := http.Serve(listener, withMW); err != nil {
		log.Fatal(err.Error())
	}
}

// newProxy takes target host and creates a reverse proxy
func newProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

// proxyRequestHandler handles the http request using proxy
func proxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
