package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/Embiggenerd/articles/core"
	"github.com/go-chi/render"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func (s *APIServer) handleGetDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		document, err := s.store.Documents.FindID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Write(document.Data.Bytes()) // create respose type
	}
}

func (s *APIServer) handleCreateDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := new(bytes.Buffer)
		_, err := io.Copy(data, r.Body)
		if err != nil {
			http.Error(w, "Failed to copy", http.StatusInternalServerError)
			return
		}
		id, err := s.store.Documents.Create(r.Context(), &core.Document{Data: *data})
		if err != nil {
			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, DocumentCreateResponse{ID: id})
		render.Status(r, http.StatusOK)
	}
}

func (s *APIServer) handleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userReq UserCreateRequest
		err := json.NewDecoder(r.Body).Decode(&userReq)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		id, err := s.store.Users.Create(r.Context(), &core.User{Email: userReq.Email, Password: userReq.Password})
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

func (s *APIServer) handleLoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user core.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Failed to read", http.StatusInternalServerError)
			return
		}

		// Check password matches hashed value stored in DB
		authenticated, foundUser, err := s.store.Users.FindEmailAndAuthenticate(r.Context(), user.Email, user.Password)
		if err != nil || !authenticated {
			http.Error(w, "Failed to authenticate", http.StatusUnauthorized)
			return
		}

		// Create a cookie token we can decrypt later
		token, err := createToken(s.cfg.Get(s.cfg.AuthSecret))
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

func (s *APIServer) handleProtected() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("exampleCookie")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "Cookie not found", http.StatusBadRequest)
			default:
				log.Println(err)
				http.Error(w, "Server error", http.StatusInternalServerError)
			}
			return
		}

		verified := verifyToken(s.cfg.Get(s.cfg.AuthSecret), cookie.Value)
		if !verified {
			http.Error(w, "Failed to authenticate", http.StatusUnauthorized)
			return
		}

		w.Write([]byte(cookie.Value))
	}
}

func createToken(secret string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(secret), 14)
	return string(bytes), err
}

func verifyToken(secret, token string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(token), []byte(secret))
	return err == nil
}
