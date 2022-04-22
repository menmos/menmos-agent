package main

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/menmos/menmos-agent/agent"
	"github.com/menmos/menmos-agent/api"
	"github.com/spf13/viper"
)

type agentConfiguration struct {
	Debug bool         `json:"debug" mapstructure:"DEBUG" toml:"debug"`
	Agent agent.Config `json:"agent" mapstructure:"AGENT" toml:"agent"`
	API   api.Config   `json:"api" mapstructure:"API" toml:"api"`
}

func defaultConfig() agentConfiguration {
	c := agentConfiguration{
		Debug: false,
		Agent: agent.Config{
			AgentType: agent.TypeNative,
			Path:      "./menmos_agent_data",
		},
		API: api.Config{
			Host: "0.0.0.0",
			Port: 3030,
		},
	}
	return c
}

func loadConfig() (agentConfiguration, error) {
	// First set the default config for viper.
	b, err := json.Marshal(defaultConfig())
	if err != nil {
		return agentConfiguration{}, err
	}
	def := bytes.NewReader(b)

	viper.SetConfigType("json")

	if err := viper.MergeConfig(def); err != nil {
		return agentConfiguration{}, err
	}

	viper.SetConfigName("agent")
	viper.AddConfigPath(".")
	if err := viper.MergeInConfig(); err != nil {
		return agentConfiguration{}, err
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("MENMOS_AGENT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config := agentConfiguration{}
	err = viper.Unmarshal(&config)
	return config, err
}
