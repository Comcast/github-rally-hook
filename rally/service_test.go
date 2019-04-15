// +build unit

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

package rally_test

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"

	"github.com/comcast/github-rally-hook/rally"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("A service connector for rally and github", func() {
	var (
		svc    rally.Service
		cfg    rally.Config
		ctx    context.Context
		server *ghttp.Server
	)

	BeforeEach(func() {
		fmt.Printf("STARTING SERVER>>>>\n")
		server = ghttp.NewServer()
		server.AllowUnhandledRequests = false
	})

	AfterEach(func() {
		fmt.Printf(">>>>CLOSING SERVER\n")
		server.Close()
	})

	Describe("ReceivePush", func() {
		var (
			pushEvent    rally.PushEvent
			pushResponse rally.PushResponse
			err          error
		)

		Context("when called with a valid github push event", func() {
			BeforeEach(func() {
				// Read in JSON files
				w, err := ioutil.ReadFile("../fixtures/success_getWorkspace.json")
				if err != nil {
					Skip(err.Error())
				}

				gs, err := ioutil.ReadFile("../fixtures/success_getSCMRepo.json")
				if err != nil {
					Skip(err.Error())
				}

				u, err := ioutil.ReadFile("../fixtures/success_getUser.json")
				if err != nil {
					Skip(err.Error())
				}

				us, err := ioutil.ReadFile("../fixtures/success_getUserStory.json")
				if err != nil {
					Skip(err.Error())
				}

				chset, err := ioutil.ReadFile("../fixtures/success_createChangeSet.json")
				if err != nil {
					Skip(err.Error())
				}

				ch, err := ioutil.ReadFile("../fixtures/success_createChange.json")
				if err != nil {
					Skip(err.Error())
				}

				pushReq, err := ioutil.ReadFile("../fixtures/sample_pushevent.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(pushReq)).Decode(&pushEvent)
				if err != nil {
					Skip(err.Error())
				}

				server.AppendHandlers(
					//Workspace get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/workspace"),
						ghttp.RespondWith(http.StatusOK, string(w[:])),
					),
					//SCMRepo Get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/scmrepository"),
						ghttp.RespondWith(http.StatusOK, string(gs[:])),
					),
					// User story get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/hierarchicalrequirement"),
						ghttp.RespondWith(http.StatusOK, string(us[:])),
					),
					// User get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/user"),
						ghttp.RespondWith(http.StatusOK, string(u[:])),
					),
					// create changeset response
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/slm/webservice/v2.0/changeset/create"),
						ghttp.RespondWith(http.StatusOK, string(chset[:])),
					),
					// create change response
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/slm/webservice/v2.0/change/create"),
						ghttp.RespondWith(http.StatusOK, string(ch[:])),
					),
				)
				cfg = rally.Config{
					RallyURL:  server.URL(),
					APIToken:  "1234abcde",
					Workspace: "Comcast",
				}
				ctx = context.Background()
				svc = rally.NewPushReceiveService(log.NewNopLogger(), cfg)
			})
			It("should create the changeset and changes without error", func() {
				pushResponse, err = svc.ReceivePush(ctx, pushEvent)
				fmt.Println(pushResponse)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
		Context("when called with a valid event and STARTS in the commit message", func() {
			BeforeEach(func() {
				// Read in JSON files
				w, err := ioutil.ReadFile("../fixtures/success_getWorkspace.json")
				if err != nil {
					Skip(err.Error())
				}

				gs, err := ioutil.ReadFile("../fixtures/success_getSCMRepo.json")
				if err != nil {
					Skip(err.Error())
				}

				u, err := ioutil.ReadFile("../fixtures/success_getUser.json")
				if err != nil {
					Skip(err.Error())
				}

				us, err := ioutil.ReadFile("../fixtures/success_getUserStory.json")
				if err != nil {
					Skip(err.Error())
				}

				chset, err := ioutil.ReadFile("../fixtures/success_createChangeSet.json")
				if err != nil {
					Skip(err.Error())
				}

				ch, err := ioutil.ReadFile("../fixtures/success_createChange.json")
				if err != nil {
					Skip(err.Error())
				}

				pushReq, err := ioutil.ReadFile("../fixtures/sample_pushevent.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(pushReq)).Decode(&pushEvent)
				if err != nil {
					Skip(err.Error())
				}

				server.AppendHandlers(
					//Workspace get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/workspace"),
						ghttp.RespondWith(http.StatusOK, string(w[:])),
					),
					//SCMRepo Get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/scmrepository"),
						ghttp.RespondWith(http.StatusOK, string(gs[:])),
					),
					// User story get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/hierarchicalrequirement"),
						ghttp.RespondWith(http.StatusOK, string(us[:])),
					),
					// User get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/user"),
						ghttp.RespondWith(http.StatusOK, string(u[:])),
					),
					// create changeset response
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/slm/webservice/v2.0/changeset/create"),
						ghttp.RespondWith(http.StatusOK, string(chset[:])),
					),
					// create change response
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/slm/webservice/v2.0/change/create"),
						ghttp.RespondWith(http.StatusOK, string(ch[:])),
					),
				)
				cfg = rally.Config{
					RallyURL:  server.URL(),
					APIToken:  "1234abcde",
					Workspace: "Comcast",
				}
				pushEvent.Commits[0].Message = "STARTS US12345 - misnamed CompletionPercentage"
				ctx = context.Background()
				svc = rally.NewPushReceiveService(log.NewNopLogger(), cfg)
			})
			It("should update the status in rally and not return an error", func() {
				pushResponse, err = svc.ReceivePush(ctx, pushEvent)
				fmt.Println(pushResponse)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
		Context("when called with a valid event and FINISHES in the commit message", func() {
			BeforeEach(func() {
				// Read in JSON files
				w, err := ioutil.ReadFile("../fixtures/success_getWorkspace.json")
				if err != nil {
					Skip(err.Error())
				}

				gs, err := ioutil.ReadFile("../fixtures/success_getSCMRepo.json")
				if err != nil {
					Skip(err.Error())
				}

				u, err := ioutil.ReadFile("../fixtures/success_getUser.json")
				if err != nil {
					Skip(err.Error())
				}

				us, err := ioutil.ReadFile("../fixtures/success_getUserStory.json")
				if err != nil {
					Skip(err.Error())
				}

				chset, err := ioutil.ReadFile("../fixtures/success_createChangeSet.json")
				if err != nil {
					Skip(err.Error())
				}

				ch, err := ioutil.ReadFile("../fixtures/success_createChange.json")
				if err != nil {
					Skip(err.Error())
				}

				pushReq, err := ioutil.ReadFile("../fixtures/sample_pushevent.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(pushReq)).Decode(&pushEvent)
				if err != nil {
					Skip(err.Error())
				}

				server.AppendHandlers(
					//Workspace get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/workspace"),
						ghttp.RespondWith(http.StatusOK, string(w[:])),
					),
					//SCMRepo Get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/scmrepository"),
						ghttp.RespondWith(http.StatusOK, string(gs[:])),
					),
					// User story get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/hierarchicalrequirement"),
						ghttp.RespondWith(http.StatusOK, string(us[:])),
					),
					// User get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/user"),
						ghttp.RespondWith(http.StatusOK, string(u[:])),
					),
					// create changeset response
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/slm/webservice/v2.0/changeset/create"),
						ghttp.RespondWith(http.StatusOK, string(chset[:])),
					),
					// create change response
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/slm/webservice/v2.0/change/create"),
						ghttp.RespondWith(http.StatusOK, string(ch[:])),
					),
				)
				cfg = rally.Config{
					RallyURL:  server.URL(),
					APIToken:  "1234abcde",
					Workspace: "Comcast",
				}
				pushEvent.Commits[0].Message = "COMPLETES US12345 - misnamed CompletionPercentage"
				ctx = context.Background()
				svc = rally.NewPushReceiveService(log.NewNopLogger(), cfg)
			})
			It("should update the status in rally and not return an error", func() {
				pushResponse, err = svc.ReceivePush(ctx, pushEvent)
				fmt.Println(pushResponse)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
	Describe(".FindRallyArtifact", func() {
		Context("when called with a commit message with more than one rally id", func() {
			var commit rally.Commit

			BeforeEach(func() {
				us, err := ioutil.ReadFile("../fixtures/success_getUserStory.json")
				if err != nil {
					Skip(err.Error())
				}
				ta, err := ioutil.ReadFile("../fixtures/success_getTask.json")
				if err != nil {
					Skip(err.Error())
				}

				cmt, err := ioutil.ReadFile("../fixtures/sample_commit_multiArtifact.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(cmt)).Decode(&commit)
				if err != nil {
					Skip(err.Error())
				}

				server.AppendHandlers(
					// User story get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/hierarchicalrequirement"),
						ghttp.RespondWith(http.StatusOK, string(us[:])),
					),
					// Task get
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/slm/webservice/v2.0/task"),
						ghttp.RespondWith(http.StatusOK, string(ta[:])),
					),
				)
				cfg = rally.Config{
					RallyURL:  server.URL(),
					APIToken:  "1234abcde",
					Workspace: "Comcast",
				}
				ctx = context.Background()
				svc = rally.NewPushReceiveService(log.NewNopLogger(), cfg)

			})
			It("should return an array of the correct number of references", func() {
				refs := svc.FindRallyArtifact(commit)
				Expect(len(refs)).Should(Equal(2))
			})
		})
		Context("when called with a commit message with no rally ids", func() {
			var commit rally.Commit

			BeforeEach(func() {
				cmt, err := ioutil.ReadFile("../fixtures/sample_commit_noArtifact.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(cmt)).Decode(&commit)
				if err != nil {
					Skip(err.Error())
				}

				cfg = rally.Config{
					RallyURL:  server.URL(),
					APIToken:  "1234abcde",
					Workspace: "Comcast",
				}
				ctx = context.Background()
				svc = rally.NewPushReceiveService(log.NewNopLogger(), cfg)

			})
			It("should return an empty array of references", func() {
				refs := svc.FindRallyArtifact(commit)
				Expect(len(refs)).Should(Equal(0))
			})
		})
	})
	Describe(".CheckHMAC", func() {

		var (
			auth        *rally.Authorizor
			secretToken string
			logger      log.Logger
			err         error
			pushEvent   rally.PushEvent
		)
		Context("when called with a valid HTTP_X_HUB_SIGNATURE in the header", func() {
			BeforeEach(func() {
				pushReq, err := ioutil.ReadFile("../fixtures/sample_pushevent.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(pushReq)).Decode(&pushEvent)
				if err != nil {
					Skip(err.Error())
				}

				if secretToken, err = randomHex(20); err != nil {
					Skip("unable to create token")
				}

				logger = log.NewNopLogger()

				signingMethod := jwt.SigningMethodHMAC{
					"SHA1",
					crypto.SHA1,
				}
				washit, err := json.Marshal(pushEvent)

				value, err := signingMethod.Sign(string(washit[:]), []byte(secretToken))
				if err != nil {
					Skip(err.Error())
				}

				fmt.Printf("Computed - %s for %s\n", value, secretToken)
				ctx = context.WithValue(context.Background(), "X-Hub-Signature", value)

				auth = &rally.Authorizor{
					SecretToken:       secretToken,
					SignatureRequired: true,
					Logger:            logger,
				}
			})
			It("should not return an error", func() {
				err = auth.CheckHMAC(ctx, pushEvent)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
		Context("when called without a valid HTTP_X_HUB_SIGNATURE in the header", func() {
			BeforeEach(func() {
				pushReq, err := ioutil.ReadFile("../fixtures/sample_pushevent.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(pushReq)).Decode(&pushEvent)
				if err != nil {
					Skip(err.Error())
				}

				logger = log.NewNopLogger()
				ctx = context.Background()

				auth = &rally.Authorizor{
					SecretToken:       secretToken,
					SignatureRequired: false,
					Logger:            logger,
				}
			})
			It("should not return an error", func() {
				err = auth.CheckHMAC(ctx, pushEvent)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
		Context("when called with an invalid HTTP_X_HUB_SIGNATURE in the header", func() {
			BeforeEach(func() {
				pushReq, err := ioutil.ReadFile("../fixtures/sample_pushevent.json")
				if err != nil {
					Skip(err.Error())
				}

				err = json.NewDecoder(bytes.NewReader(pushReq)).Decode(&pushEvent)
				if err != nil {
					Skip(err.Error())
				}

				if secretToken, err = randomHex(20); err != nil {
					Skip("unable to create token")
				}

				logger = log.NewNopLogger()
				ctx = context.WithValue(context.Background(), "X-Hub-Signature", "somerandostring")

				auth = &rally.Authorizor{
					SecretToken:       "secretoken",
					SignatureRequired: true,
					Logger:            logger,
				}
			})
			It("should return an error", func() {
				err = auth.CheckHMAC(ctx, pushEvent)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
})

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func computeHmac1(message []byte, secret []byte) string {
	h := hmac.New(sha1.New, secret)
	h.Write(message)
	fmt.Println(h.Sum(nil))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
