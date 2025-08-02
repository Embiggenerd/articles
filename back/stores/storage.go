package stores

import (
	"os"

	"github.com/Embiggenerd/articles/core"

	"github.com/Embiggenerd/articles/stores/sqlite"

	"github.com/sirupsen/logrus"
)

type Store struct {
	Documents core.DocumentStore
	Users     core.UserStore
}

func GetStore() Store {
	storageType := os.Getenv("STORAGE_TYPE")
	store := Store{}

	storageField := logrus.Fields{
		"storageType": storageType,
	}

	switch storageType {
	// case "filesystem":
	// 	basePath := os.Getenv("LOCAL_STORAGE_PATH")
	// 	storageField["basePath"] = basePath
	// 	store = filesystem.NewDocumentStore(basePath)
	case "sqlite":
		// dataSourceName := os.Getenv("DATA_SOURCE_NAME")
		dataSourceName := "test.sqlite"
		storageField["dataSourceName"] = dataSourceName
		store.Documents = sqlite.NewDocumentStore(dataSourceName)
		store.Users = sqlite.NewUserStore(dataSourceName)
		// case "s3":
		// 	bucketName := os.Getenv("S3_BUCKET_NAME")
		// 	storageField["bucketName"] = bucketName
		// 	store = aws.NewDocumentStore(bucketName)
		// default:
		// 	store = memory.NewDocumentStore()
		// 	storageField["storageType"] = "in-memory"
	}
	logrus.WithFields(storageField).Info("Use storage")
	return store
}
