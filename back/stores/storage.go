package stores

import (
	"context"

	"github.com/Embiggenerd/articles/config"
	"github.com/Embiggenerd/articles/core"
	"github.com/Embiggenerd/articles/logger"

	"github.com/Embiggenerd/articles/stores/sqlite"

	"github.com/sirupsen/logrus"
)

type Store struct {
	Documents core.DocumentStore
	Users     core.UserStore
}

func GetStores(ctx context.Context, cfg *config.Config, log logger.Logger) Store {
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
	log.Info("Use storage", storageField)
	log.Info("Use storage", storageField)
	return store
}
