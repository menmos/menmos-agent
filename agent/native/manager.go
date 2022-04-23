package native

import (
	"github.com/menmos/menmos-agent/agent/native/artifact"
	"go.uber.org/zap"
)

type NativeManager struct {
	config NativeAgentConfig
	log    *zap.SugaredLogger

	artifacts *artifact.Repository
}

func (m *NativeManager) CreateMenmosd() {}
