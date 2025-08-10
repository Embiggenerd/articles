package stores

import (
	"github.com/Embiggenerd/articles/config"
	"github.com/Embiggenerd/articles/core"

	"github.com/Embiggenerd/articles/stores/sqlite"

	"github.com/sirupsen/logrus"
)

type Store struct {
	Documents core.DocumentStore
	Users     core.UserStore
}

func GetStore(cfg config.Config) Store {
	storageType := cfg.Get(cfg.StorageType)
	store := Store{}

	storageField := logrus.Fields{
		"storageType": storageType,
	}

	switch storageType {
	case "sqlite":
		dataSourceName := cfg.Get(cfg.DataSourceName)
		storageField["dataSourceName"] = dataSourceName
		store.Documents = sqlite.NewDocumentStore(dataSourceName)
		store.Users = sqlite.NewUserStore(dataSourceName)
	}
	logrus.WithFields(storageField).Info("Use storage")
	return store
}
