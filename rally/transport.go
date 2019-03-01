/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package rally

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
)

var (
	// ErrBadRouting - returns 400 http error
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
	// ErrInvalidArgument - returns 400 http error
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrUnauthorized - returns 401 http error
	ErrUnauthorized = errors.New("authorization is invalid")
	// ErrInvalidToken - returns 401 http error
	ErrInvalidToken = errors.New("token contains an invalid number of segments")
	// ErrForbidden - returns 403 http error
	ErrForbidden = errors.New("User not authorized for operation")
)

// MakeRoutes - make routes
func MakeRoutes(r *mux.Router, s Service, logger log.Logger, middleware endpoint.Middleware, auth ...kithttp.RequestFunc) {
	options := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(auth...),
	}

	r.Methods("POST").Path("/receive").Handler(kithttp.NewServer(
		middleware(MakePushEventEndpoint(s)),
		decodePushEventRequest,
		encodeResponse,
		options...,
	))
}

func decodePushEventRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var event PushEvent

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		return nil, ErrInvalidArgument
	}
	return event, nil
}
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError

	switch err {
	case ErrInvalidArgument:
		code = http.StatusBadRequest
	case ErrUnauthorized:
		code = http.StatusUnauthorized
	case ErrForbidden:
		code = http.StatusForbidden
	case ErrInvalidToken:
		code = http.StatusUnauthorized
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
