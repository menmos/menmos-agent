package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/menmos/menmos-agent/agent"
	"github.com/menmos/menmos-agent/api"
	"go.uber.org/zap"
)

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	// Setup logging.
	var logger *zap.Logger
	if config.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}

	defer logger.Sync()

	agt, err := agent.New(config.Agent, logger)
	if err != nil {
		panic(err)
	}

	srv := api.New(agt, config.API, logger)
	if err := srv.Start(); err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	agt.Shutdown()
	os.Exit(0)

}
