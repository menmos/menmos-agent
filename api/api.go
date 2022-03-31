package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/menmos/menmos-agent/agent"
	"github.com/menmos/menmos-agent/payload"
	"go.uber.org/zap"
)

// API regroups the route of the agent API.
type API struct {
	agent  *agent.MenmosAgent
	config Config
	log    *zap.SugaredLogger
}

// New returns a new API instance.
func New(agent *agent.MenmosAgent, config Config, log *zap.Logger) *API {
	return &API{
		agent:  agent,
		config: config,
		log:    log.Sugar().Named("api"),
	}
}

func (a *API) healthCheck(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return payload.HealthCheckResponse{Status: "healthy"}, nil
}

func (a *API) serve() {
	r := mux.NewRouter()

	r.HandleFunc("/health", wrapRoute(a.log, a.healthCheck))

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", a.config.Host, a.config.Port), r))
}

// Start starts the API server.
func (a *API) Start() error {
	a.log.Info("server starting")
	go a.serve()
	a.log.Infof("server started on [%s:%d]", a.config.Host, a.config.Port)
	return nil
}
