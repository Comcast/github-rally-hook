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
	"crypto"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

type Authorizor struct {
	SecretToken       string
	SignatureRequired bool
	Logger            log.Logger
}

func (a *Authorizor) ValidatePayload() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			if err := a.CheckHMAC(ctx, request); err != nil {
				return nil, err
			}

			return next(ctx, request)
		}
	}
}

func (a *Authorizor) CheckHMAC(ctx context.Context, request interface{}) (err error) {
	logger := log.With(a.Logger, "event", "CheckHMAC")

	signature, ok := ctx.Value("X-Hub-Signature").(string)

	// If the signature is not present and not required then bypass
	if !ok && !a.SignatureRequired {
		logger.Log("message", "signature not present")
		return nil
	}

	signingMethod := jwt.SigningMethodHMAC{
		"SHA1",
		crypto.SHA1,
	}

	requestBytes, err := json.Marshal(request.(PushEvent))

	if err != nil {
		return err
	}
	if err = signingMethod.Verify(string(requestBytes[:]), signature, []byte(a.SecretToken)); err != nil {
		return err
	}

	return nil
}
