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
	"time"

	"context"
	"github.com/go-kit/kit/log"
)

// NewLoggingService logs each method call to s.
func NewLoggingService(s Service, logger log.Logger) Service {
	return &loggingService{
		s:      s,
		logger: logger,
	}
}

type loggingService struct {
	s      Service
	logger log.Logger
}

func (l *loggingService) ReceivePush(ctx context.Context, request PushEvent) (response PushResponse, err error) {
	defer func(start time.Time) {
		l.logger.Log("event", "ReceivePush", "err", err, "dur", time.Since(start))
	}(time.Now())
	return l.s.ReceivePush(ctx, request)
}

func (l *loggingService) FindRallyArtifact(commit Commit) (artifacts map[string]string) {
	defer func(start time.Time) {
		l.logger.Log("event", "FindArtifacts", "dur", time.Since(start))
	}(time.Now())
	return l.s.FindRallyArtifact(commit)
}
