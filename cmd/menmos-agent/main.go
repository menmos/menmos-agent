package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/menmos/menmos-agent/agent"
	"github.com/menmos/menmos-agent/api"
	"go.uber.org/zap"
)

const MENMOSD_PATH = "/Users/wduss/src/github.com/menmos/menmos/target/debug/menmosd"

func main() {

	// Setup logging.
	//logger, err := zap.NewProduction()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}

	defer logger.Sync()

	config := agent.Config{AgentType: agent.Native, Path: "./agent_workspace"}
	agt, err := agent.New(config, logger)
	if err != nil {
		panic(err)
	}

	srv := api.New(agt, api.Config{Host: "0.0.0.0", Port: 3030}, logger)
	if err := srv.Start(); err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	agt.Shutdown()
	os.Exit(0)

}
