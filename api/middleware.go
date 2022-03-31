package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type errorResponse struct {
	Error string `json:"error,omitempty"`
}

func logStatus(log *zap.SugaredLogger, r *http.Request, status int) {
	if log == nil {
		return
	}

	log.Infof("[%s] %s - [%d]", r.Method, r.URL.Path, status)
}

func handleError(w http.ResponseWriter, err error, r *http.Request, log *zap.SugaredLogger) {

	body, jsonErr := json.Marshal(&errorResponse{Error: err.Error()})
	if jsonErr != nil {
		panic(jsonErr)
	}

	statusCode := http.StatusInternalServerError
	if errors.Is(err, errInternalServerError) {
		log.Errorf("error processing request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if errors.Is(err, errBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
	} else if errors.Is(err, errNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		log.Errorf("unhandled error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(body)
	logStatus(log, r, statusCode)
}

func wrapRoute(log *zap.SugaredLogger, f func(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		rval, err := f(ctx, w, r)
		w.Header().Add("Content-Type", "aplication/json")

		if err != nil {
			handleError(w, err, r, log)
		} else {
			raw, err := json.Marshal(rval)
			if err != nil {
				handleError(w, err, r, log)
			}
			w.Write(raw)
			logStatus(log, r, http.StatusOK)
		}
	}
}
