package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type errorResponse struct {
	Error error `json:"error,omitempty"`
}

func logStatus(log *zap.SugaredLogger, r *http.Request, status int) {
	fmt.Println("logging status")
	if log == nil {
		fmt.Println("short-circuit")
		return
	}

	str := fmt.Sprintf("[%s] %s - [%d]", r.Method, r.URL.Path, status)
	if status == http.StatusOK {
		fmt.Println("logging ok")
		log.Info(str)
	} else {
		log.Warn(str)
	}
}

func handleError(w http.ResponseWriter, err error, r *http.Request, log *zap.SugaredLogger) {

	body, err := json.Marshal(&errorResponse{Error: err})
	if err != nil {
		panic(err)
	}

	statusCode := http.StatusInternalServerError
	if errors.Is(err, errInternalServerError) {
		log.Errorf("error processing request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Errorf("unhandled error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(body)
	logStatus(log, r, statusCode)
}

func wrapRoute(log *zap.SugaredLogger, f func(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("here in middleware")
		ctx := context.Background()

		rval, err := f(ctx, w, r)
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
