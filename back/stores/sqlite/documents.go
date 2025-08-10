package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Embiggenerd/articles/core"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

type documentStore struct {
	db *sql.DB
}

func NewDocumentStore(dataSourceName string) core.DocumentStore {
	// db, err := sql.Open("sqlite3", ":memory:")
	dirPath := "../data" // Path to the directory you want to create

	// // Define the permissions for the new directory (e.g., 0755 for rwxr-xr-x)
	// // os.ModePerm is 0777, granting full permissions.
	// // You can use specific octal values like 0755 for common permissions.
	permissions := os.ModePerm

	err := os.MkdirAll(dirPath, permissions)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		// return
	}
	db, err := sql.Open("sqlite3", dirPath+"/"+dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	sts := `CREATE TABLE IF NOT EXISTS documents (id TEXT PRIMARY KEY, data BLOB);`
	_, err = db.Exec(sts)
	if err != nil {
		log.Fatal(err)
	}
	return &documentStore{db}
}

func (s *documentStore) FindID(ctx context.Context, id string) (*core.Document, error) {
	log := logrus.WithField("document_id", id)
	log.Debug("Retrieving document by ID")
	var data []byte
	err := s.db.QueryRowContext(ctx, "SELECT data FROM documents WHERE id = ?", id).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithField("error", "document not found").Warn("Document with specified ID not found")
			return nil, fmt.Errorf("document with id %s not found", id)
		}
		log.WithField("error", err).Error("Failed to retrieve document")
		return nil, err
	}
	document := core.Document{
		Data: *bytes.NewBuffer(data),
	}
	log.Info("Document retrieved successfully")
	return &document, nil
}

func (s *documentStore) Create(ctx context.Context, document *core.Document) (string, error) {
	id := ulid.Make().String()
	data := document.Data.Bytes()
	log := logrus.WithFields(logrus.Fields{
		"document_id": id,
		"data_length": len(data),
	})

	_, err := s.db.ExecContext(ctx, "INSERT INTO documents (id, data) VALUES (?, ?)", id, data)
	if err != nil {
		log.WithField("error", err).Error("Failed to create document")
		return "", err
	}
	log.Info("Document created successfully")
	return id, nil
}
