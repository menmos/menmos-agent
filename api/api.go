package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/menmos/menmos-agent/agent"
	"github.com/menmos/menmos-agent/payload"
	"go.uber.org/zap"
)

// API regroups the route of the agent API.
type API struct {
	agent  *agent.Agent
	config Config
	log    *zap.SugaredLogger
}

// New returns a new API instance.
func New(agent *agent.Agent, config Config, log *zap.Logger) *API {
	return &API{
		agent:  agent,
		config: config,
		log:    log.Sugar().Named("api"),
	}
}

func (a *API) healthCheck(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return payload.HealthCheckResponse{Status: "healthy"}, nil
}

func (a *API) createNode(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var request payload.CreateNodeRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &request); err != nil {
		return nil, errBadRequest
	}

	return a.agent.CreateNode(&request)
}

func (a *API) listNodes(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return a.agent.ListNodes()
}

func (a *API) getNode(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		node, err := a.agent.GetNode(id)
		if err != nil {
			return nil, err
		}

		if node == nil {
			return nil, errNotFound
		}

		return node, nil
	}
	panic("bad routing config")
}

func (a *API) deleteNode(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		if err := a.agent.DeleteNode(id); err != nil {
			return nil, err
		}
		return payload.MessageResponse{Message: "ok"}, nil
	}
	panic("bad routing config")
}

func (a *API) getNodeLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		logs, err := a.agent.GetNodeLogs(id, 30) // TODO: take query param here
		return logs, err
	}
	panic("bad routing config")
}

func (a *API) startNode(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		if err := a.agent.StartNode(id); err != nil {
			return nil, err
		}

		return payload.MessageResponse{Message: "ok"}, nil
	}
	panic("bad routing config")

}

func (a *API) stopNode(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		if err := a.agent.StopNode(id); err != nil {
			return nil, err
		}

		return payload.MessageResponse{Message: "ok"}, nil
	}
	panic("bad routing config")

}

func (a *API) serve() {
	r := mux.NewRouter()

	// Node CRUD.
	r.HandleFunc("/node", wrapRoute(a.log, a.createNode)).Methods("POST")
	r.HandleFunc("/node", wrapRoute(a.log, a.listNodes)).Methods("GET")
	r.HandleFunc("/node/{id}", wrapRoute(a.log, a.getNode)).Methods("GET")
	r.HandleFunc("/node/{id}", wrapRoute(a.log, a.deleteNode)).Methods("DELETE")
	r.HandleFunc("/node/{id}/logs", wrapRoute(a.log, a.getNodeLogs)).Methods("GET")
	r.HandleFunc("/node/{id}/start", wrapRoute(a.log, a.startNode)).Methods("POST")
	r.HandleFunc("/node/{id}/stop", wrapRoute(a.log, a.stopNode)).Methods("POST")

	// Misc.
	r.HandleFunc("/health", wrapRoute(a.log, a.healthCheck)).Methods("GET")

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", a.config.Host, a.config.Port), r))
}

// Start starts the API server.
func (a *API) Start() error {
	a.log.Info("server starting")
	go a.serve()
	a.log.Infof("server started on [%s:%d]", a.config.Host, a.config.Port)
	return nil
}
