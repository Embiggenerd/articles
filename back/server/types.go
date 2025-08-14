package server

import "github.com/Embiggenerd/articles/core"

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
