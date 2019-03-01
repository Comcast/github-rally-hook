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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/comcast/github-rally-hook/rally"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	kitinflux "github.com/go-kit/kit/metrics/influx"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/influxdb/client/v2"

	"github.com/go-stack/stack"
)

func main() {

	newLogger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger := newLogContext(newLogger, "API")
	metricsLogger := newLogContext(newLogger, "Metrics")
	pushLogger := newLogContext(newLogger, "push")

	//Load the configuration from file
	logger.Log("event", "loadingConfig")

	var cfg rally.Config
	filename := "config.json"
	if _, err := os.Stat(filename); err == nil {
		f, _ := ioutil.ReadFile(filename)
		if err != nil {
			logger.Log("event", "exiting", "err", err)
			os.Exit(1)
		}
		err = json.Unmarshal(f, &cfg)
		if err != nil {
			logger.Log("event", "exiting", "err", err)
			os.Exit(1)
		}
	} else {
		logger.Log("event", "exiting", "err", err)
		os.Exit(1)
	}

	auth := &rally.Authorizor{
		SecretToken:       cfg.SecretToken,
		SignatureRequired: cfg.SignatureRequired,
		Logger:            logger,
	}

	middleware := auth.ValidatePayload()

	authBefore := []kithttp.RequestFunc{
		HTTPToContext(),
	}

	receiveService := rally.NewPushReceiveService(pushLogger, cfg)
	receiveService = rally.NewLoggingService(receiveService, logger)

	// Make the metrics optional based on whether config contains
	if cfg.InfluxCfg.URL != "" {
		in := kitinflux.New(
			map[string]string{
				"svc": "github-rally-hook",
				"env": cfg.InfluxCfg.Tag,
			},
			client.BatchPointsConfig{
				Database:        cfg.InfluxCfg.Database,
				Precision:       "ms",
				RetentionPolicy: "",
			}, metricsLogger)

		//influxdb connection
		requestCounter := in.NewCounter("requests")
		callDur := in.NewHistogram("callDur")

		client, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     cfg.InfluxCfg.URL,
			Username: cfg.InfluxCfg.Username,
			Password: cfg.InfluxCfg.Password,
		})

		if err != nil {
			logger.Log("event", "exiting", "err", err)
			os.Exit(1)
		}
		//Our Ticker Code For Batching
		ticker := time.NewTicker(5 * time.Second)

		//Our Writeloop for Batching using ticker.C channel data
		go in.WriteLoop(ticker.C, client)

		receiveService = rally.NewInstrumentedService(receiveService, requestCounter, callDur, client, in)
	}

	//Set up and start http server
	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api").Subrouter()
	rally.MakeRoutes(apiRouter, receiveService, logger, middleware, authBefore...)

	port := os.Getenv("PORT")

	if port == "" {
		logger.Log("event", "exiting", "err", "environment variable port not set")
		os.Exit(1)
	}

	server := http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	err := server.ListenAndServe()

	if err != nil {
		logger.Log("event", "exiting", "err", err)
		os.Exit(1)
	}
}

func newLogContext(logger log.Logger, app string) log.Logger {
	return log.With(logger,
		"time", log.DefaultTimestampUTC,
		"app", app,
		"caller", log.Valuer(func() interface{} {
			return pkgCaller{stack.Caller(3)}
		}),
	)

}

// pkgCaller wraps a stack.Call to make the default string output include the
// package path.
type pkgCaller struct {
	c stack.Call
}

func (pc pkgCaller) String() string {
	caller := fmt.Sprintf("%+v", pc.c)
	caller = strings.TrimPrefix(caller, "github.com/comcast/github-rally-hook/")
	return caller
}

// HTTPToContext - used to move the Github signature from the header to the context.
func HTTPToContext() kithttp.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		token := r.Header.Get("X-Hub-Signature")
		if len(token) == 0 {
			return ctx
		}

		return context.WithValue(ctx, "X-Hub-Signature", token)
	}
}
