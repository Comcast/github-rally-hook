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
	"github.com/go-kit/kit/metrics"
	kitinflux "github.com/go-kit/kit/metrics/influx"
	"github.com/influxdata/influxdb/client/v2"
)

// NewInstrumentedService - contructor function to wrap Service for metrics
func NewInstrumentedService(s Service, count metrics.Counter, callDur metrics.Histogram, c client.Client, in *kitinflux.Influx) Service {
	return &instrumentedService{
		s:       s,
		count:   count,
		callDur: callDur,
		c:       c,
		in:      in,
	}
}

type instrumentedService struct {
	s       Service
	count   metrics.Counter
	callDur metrics.Histogram
	c       client.Client
	in      *kitinflux.Influx
}

func (i *instrumentedService) ReceivePush(ctx context.Context, request PushEvent) (PushResponse, error) {
	counter := i.count.With("method", "ReceivePush")
	timer := metrics.NewTimer(i.callDur.With("method", "ReceivePush"))

	defer func() {
		counter.Add(1)
		timer.ObserveDuration()
	}()

	return i.s.ReceivePush(ctx, request)
}

func (i *instrumentedService) FindRallyArtifact(commit Commit) (artifacts map[string]string) {
	counter := i.count.With("method", "FindRallyArtifact")
	timer := metrics.NewTimer(i.callDur.With("method", "FindRallyArtifact"))

	defer func() {
		counter.Add(1)
		timer.ObserveDuration()
	}()

	return i.s.FindRallyArtifact(commit)
}
