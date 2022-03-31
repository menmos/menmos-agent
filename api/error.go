package api

import "errors"

var errBadRequest = errors.New("bad request")
var errInternalServerError = errors.New("internal server error")
var errNotFound = errors.New("not found")
