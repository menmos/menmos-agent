package main

import (
	"log"
	"os"

	"github.com/menmos/menmos-agent/cmd/menmos-platform-cli/action/agent"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "agent",
				Aliases: []string{"a"},
				Usage:   "agent management",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "host",
						Usage:    "the agent host",
						EnvVars:  []string{"MENMOS_PLATFORM_AGENT"},
						Required: true,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:   "list-nodes",
						Usage:  "list nodes running on an agent",
						Action: agent.List,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
