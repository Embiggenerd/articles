package core

import (
	"bytes"
	"context"
)

type (
	Document struct {
		Data bytes.Buffer
	}

	DocumentStore interface {
		FindID(ctx context.Context, id string) (*Document, error)
		Create(ctx context.Context, document *Document) (string, error)
	}
)

type (
	User struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}

	UserStore interface {
		FindID(ctx context.Context, id string) (*User, error)
		FindEmail(ctx context.Context, email string) (*User, error)
		Create(ctx context.Context, user *User) (string, error)
		FindEmailAndAuthenticate(ctx context.Context, email, password string) (bool, *User, error)
	}
)
