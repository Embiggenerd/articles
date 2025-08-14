package main

import (
	"context"
	"fmt"

	"github.com/Embiggenerd/articles/config"
	"github.com/Embiggenerd/articles/logger"
	"github.com/Embiggenerd/articles/server"
	"github.com/Embiggenerd/articles/stores"
)

func main() {
	fmt.Println("Starting server")
	// ctx, cancel := context.WithCancel(logger.WithMetadata(context.Background()))
	// defer cancel()
	ctx := context.Background()

	cfg := config.LoadConfig()
	fmt.Println("Config", cfg)
	log := logger.NewLogger(ctx, cfg)
	fmt.Println("Getting stores", log)
	store := stores.GetStores(ctx, cfg, log)
	server := server.NewServer(ctx, cfg, log, store)
	server.Run()
}
