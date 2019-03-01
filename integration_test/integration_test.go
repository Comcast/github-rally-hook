// +build integration

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

package integration_test

import (
	"bytes"
	"crypto"
	"github.com/dgrijalva/jwt-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"time"

	"github.com/comcast/github-rally-hook/rally"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	TargetURL   = "TARGET_URL"
	SecretToken = "SECRET_TOKEN"
)

var _ = Describe("Rally Github Service Integration Test", func() {
	var (
		client      *http.Client
		targetURL   string
		secretToken string
		request     *http.Request
		response    *http.Response
		err         error
		pushEvent   rally.PushEvent
	)

	BeforeEach(func() {
		targetURL = os.Getenv(TargetURL)
		if targetURL == "" {
			Skip("Target URL not set")
		}
		secretToken = os.Getenv(SecretToken)
		if secretToken == "" {
			Skip("Secret Token not set")
		}
	})

	Context("when called with a valid push request payload", func() {
		BeforeEach(func() {
			pushReq, err := ioutil.ReadFile("../fixtures/push_event.json")
			if err != nil {
				Skip(err.Error())
			}

			//Unmarshall json into struct and replace story ID
			err = json.Unmarshal(pushReq, &pushEvent)
			if err != nil {
				Skip(err.Error())
			}

			if len(StoryID) == 0 {
				Skip("story id is not set")
			}
			pushEvent.Commits[0].Message = fmt.Sprintf("%s - This is a test commmit message", StoryID)

			bodyBytes, err := json.Marshal(pushEvent)
			if err != nil {
				Skip(err.Error())
			}

			client = &http.Client{}

			request, err = http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
			if err != nil {
				Skip(err.Error())
			}

		})
		It("should process the event and the changeset should be added to the story", func() {
			response, err = client.Do(request)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.StatusCode).Should(Equal(200))

			// Sleep 20s to allow the goroutines making the push to rally to complete
			time.Sleep(20 * time.Second)
			changeCount, _ := GetChangeSetCount(APIKey, StoryID)
			Expect(changeCount).Should(Equal(len(pushEvent.Commits)))
		})
	})
	Context("when called with a valid signed payload", func() {
		BeforeEach(func() {
			pushReq, err := ioutil.ReadFile("../fixtures/push_event.json")
			if err != nil {
				Skip(err.Error())
			}

			//Unmarshall json into struct and replace story ID
			err = json.Unmarshal(pushReq, &pushEvent)
			if err != nil {
				Skip(err.Error())
			}

			if len(StoryID) == 0 {
				Skip("story id is not set")
			}
			pushEvent.Commits[0].Message = fmt.Sprintf("%s - This is a test commmit message", StoryID)

			bodyBytes, err := json.Marshal(pushEvent)
			if err != nil {
				Skip(err.Error())
			}
			signingMethod := jwt.SigningMethodHMAC{
				"SHA1",
				crypto.SHA1,
			}

			value, err := signingMethod.Sign(string(bodyBytes[:]), []byte(secretToken))
			if err != nil {
				Skip(err.Error())
			}
			client = &http.Client{}

			request, err = http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
			if err != nil {
				Skip(err.Error())
			}
			request.Header.Set("X-Hub-Signature", value)

		})
		It("should process the event and validate the payload", func() {
			response, err = client.Do(request)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.StatusCode).Should(Equal(200))

			// Sleep 20s to allow the goroutines making the push to rally to complete
			time.Sleep(20 * time.Second)
			changeCount, _ := GetChangeSetCount(APIKey, StoryID)
			Expect(changeCount).Should(Equal(2))
		})
	})
	Context("when called with an invalid signed payload", func() {
		BeforeEach(func() {
			pushReq, err := ioutil.ReadFile("../fixtures/push_event.json")
			if err != nil {
				Skip(err.Error())
			}

			//Unmarshall json into struct and replace story ID
			err = json.Unmarshal(pushReq, &pushEvent)
			if err != nil {
				Skip(err.Error())
			}

			if len(StoryID) == 0 {
				Skip("story id is not set")
			}
			pushEvent.Commits[0].Message = fmt.Sprintf("%s - This is a test commmit message", StoryID)

			bodyBytes, err := json.Marshal(pushEvent)
			if err != nil {
				Skip(err.Error())
			}

			client = &http.Client{}

			request, err = http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
			if err != nil {
				Skip(err.Error())
			}
			request.Header.Set("X-Hub-Signature", "somerandomvalue")

		})
		It("should reject the event and return an error", func() {
			response, err = client.Do(request)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.StatusCode).Should(Equal(500))
		})
	})

})
