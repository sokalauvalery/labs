package main

import (
	"context"
	"os"
	"time"
)

// TODO: it said that source type can be changed in future
// consider to move it in configuration file
type sourceType string

const (
	game       sourceType = "game"
	serverType sourceType = "server"
	payment    sourceType = "payment"

	sourceTypeHeaderName = "Source-Type"
)

var sourceTypes = map[sourceType]interface{}{game: nil, serverType: nil, payment: nil}

func updateState(request updateRequest, sType sourceType) error {
	Log("start update state processing source-type: %v", sType)
	return nil
}

type appConfig struct {
	DB            DbConfig
	Server        ServerConf
	PostProcessor PostProcessorConfig
}

func main() {
	// TODO: add confituration system (via yaml file / environment / commandline )
	cfg := appConfig{
		DB: DbConfig{
			Conn:   ":memory:",
			Driver: "sqlite3",
		},
		Server: ServerConf{
			Port: "8090",
		},
		PostProcessor: PostProcessorConfig{
			Interval: 60 * time.Second,
			Count:    4,
		},
	}

	db, err := NewDB(cfg.DB)
	if err != nil {
		Log("failed to open database connection %v", err)
		os.Exit(1)
	}

	mgr, err := NewStateManager(db)
	if err != nil {
		Log("failed to initialize state manager %v", err)
		os.Exit(1)
	}

	srv, err := NewServer(mgr, cfg.Server)
	if err != nil {
		Log("failed to start http server %v", err)
		os.Exit(1)
	}

	ctx, cancelContext := context.WithCancel(context.Background())
	defer (func() {
		cancelContext()
	})()

	postProcessor := NewPostProcessor(ctx, cfg.PostProcessor.Count, cfg.PostProcessor.Interval, mgr)

	go postProcessor.Run()
	srv.Run()
}
