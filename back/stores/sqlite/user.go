package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Embiggenerd/articles/core"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

var savedUser = make(map[string]core.User)

type userStore struct {
	db *sql.DB
}

func NewUserStore(dataSourceName string) core.UserStore {
	// db, err := sql.Open("sqlite3", ":memory:")
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	sts := `CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, username TEXT, password TEXT, email TEXT);`
	_, err = db.Exec(sts)
	if err != nil {
		log.Fatal(err)
	}
	return &userStore{db}
}

func (s *userStore) FindID(ctx context.Context, id string) (*core.User, error) {
	log := logrus.WithField("user_id", id)
	log.Debug("Retrieving user by ID")
	var user core.User
	err := s.db.QueryRowContext(ctx, "SELECT data FROM users WHERE id = ?", id).Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithField("error", "user not found").Warn("user with specified ID not found")
			return nil, fmt.Errorf("user with id %s not found", id)
		}
		log.WithField("error", err).Error("Failed to retrieve user")
		return nil, err
	}
	// user := core.User{
	// 	Data: *bytes.NewBuffer(data),
	// }
	log.Info("user retrieved successfully")
	return &user, nil
}

func (s *userStore) FindEmail(ctx context.Context, email string) (*core.User, error) {
	log := logrus.WithField("user_email", email)
	log.Debug("Retrieving user by Email")
	var user core.User
	err := s.db.QueryRowContext(ctx, "SELECT id, email, username, password FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.UserName, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithField("error", "user not found").Warn("user with specified email not found")
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		log.WithField("error", err).Error("Failed to retrieve user")
		return nil, err
	}
	// user := core.User{
	// 	Data: *bytes.NewBuffer(data),
	// }
	log.Info("user retrieved successfully")
	return &user, nil
}

func (s *userStore) Create(ctx context.Context, user *core.User) (string, error) {
	user.ID = ulid.Make().String()
	hashedPass, err := hashPassword(user.Password)

	log := logrus.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"user_email": user.Email,
		"user_name":  user.UserName,
	})

	if err != nil {
		log.WithField("error", err).Error("Failed to create user")
		return "", err
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)", user.ID, user.UserName, user.Email, hashedPass)
	if err != nil {
		log.WithField("error", err).Error("Failed to create user")
		return "", err
	}
	log.Info("user created successfully")
	return user.ID, nil
}

func (s *userStore) FindEmailAndAuthenticate(ctx context.Context, email, password string) (bool, *core.User, error) {
	log := logrus.WithFields(logrus.Fields{
		"user_email": email,
	})
	user, err := s.FindEmail(ctx, email)
	if err != nil {
		log.WithField("error", err).Error("Failed to find user by email")
		return false, nil, err
	}

	return verifyPassword(password, user.Password), user, err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
