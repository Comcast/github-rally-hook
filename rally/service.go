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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type Service interface {
	ReceivePush(ctx context.Context, event PushEvent) (PushResponse, error)
	FindRallyArtifact(commit Commit) (artifacts map[string]string)
}

type service struct {
	logger log.Logger
	mut    sync.Mutex
	cfg    Config
	client *http.Client
}

var userCache map[string]string

func NewPushReceiveService(l log.Logger, cfg Config) Service {
	return &service{
		logger: l,
		cfg:    cfg,
		client: &http.Client{},
	}
}

func (s *service) ReceivePush(ctx context.Context, event PushEvent) (response PushResponse, err error) {

	var (
		branch  string
		repo    string
		repoURL string
	)

	logger := log.With(s.logger, "event", "ReceivePush")
	splitRef := strings.Split(event.Ref, "/")

	if len(splitRef) > 0 {
		branch = splitRef[len(splitRef)-1]
	}

	repo = event.Repository.Name
	repoURL = event.Repository.URL

	logger.Log("repo", repo, "repoURL", repoURL, "branch", branch)

	workspaceRef, ok := s.ValidateOrg(s.cfg.Workspace)
	if !ok {
		return PushResponse{Result: "workspace not found"}, errors.New("workspace not found")
	}

	// Large commits can cause Github to timeout and drop the transaction, spinning off to a goroutine allows the process to complete asynchronously
	go func(ev PushEvent, workspaceRef string) {
		// Get or Create Rally SCM repo
		scmrepo, err := s.GetOrCreateSCMRepository(repo, repoURL, workspaceRef)

		if err != nil {
			logger.Log("GetOrCreateSCMRepository", repo, "err", err.Error())
		}
		userCache = make(map[string]string)

		// For each commit extract the rally ID and add a changeset
		// Create a map of formatted id's to references
		for _, c := range event.Commits {
			refs := s.FindRallyArtifact(c)
			s.AddChangeSet(c, scmrepo, refs, repoURL, branch)
		}
		logger.Log("status", "Update rally completed")
	}(event, workspaceRef)

	return PushResponse{Result: "created"}, nil
}

func (s *service) AddChangeSet(c Commit, scmrepo string, rallyRef map[string]string, repoURL string, branch string) error {
	author := c.Author.Email

	var err error

	// Cache the user ref so not look up each time
	if _, ok := userCache[author]; !ok {
		urlString := fmt.Sprintf("%s/slm/webservice/v2.0/user", s.cfg.RallyURL)
		req, _ := http.NewRequest(http.MethodGet, urlString, nil)

		params := url.Values{}
		params.Set("query", fmt.Sprintf("(UserName = %s)", author))

		req.URL.RawQuery = params.Encode()

		s.DecorateRequest(req)

		var rallyresponse RallyQueryResults

		response, err := s.client.Do(req)

		if err == nil {
			defer response.Body.Close()
			if err = json.NewDecoder(response.Body).Decode(&rallyresponse); err == nil {
				results := rallyresponse.QueryResult.Results
				if len(results) == 1 {
					userCache[author] = results[0].Ref
				} else {
					userCache[author] = ""
				}
			} else {
				userCache[author] = ""
			}
		} else {
			userCache[author] = ""
		}
	}

	var artifactRefs []Reference

	if len(rallyRef) > 0 {
		for k, v := range rallyRef {
			artifactRefs = append(artifactRefs, Reference{Ref: v})

			//For each of the artifact references check and update scheduled state as required
			starts, completes := s.checkForStatus(c.Message, k)

			state := ""
			if starts {
				state = "In-Progress"
			}

			if completes {
				state = "Completed"
			}

			if state != "" {
				if err := s.UpdateState(v, state); err != nil {
					fmt.Printf("Error updating state: %s", err.Error())
				}
			}
		}
	}
	// Create a changeset
	userRef := userCache[author]
	changeSet := Changeset{
		SCMRepository:   scmrepo,
		Revision:        c.ID,
		Message:         c.Message,
		Uri:             fmt.Sprintf("%s/commit/%s", repoURL, c.ID),
		CommitTimestamp: c.Timestamp,
	}

	if len(artifactRefs) > 0 {
		changeSet.Artifacts = artifactRefs
	}

	if userRef != "" {
		changeSet.Author = userRef
	}

	createBody := map[string]interface{}{
		"Changeset": changeSet,
	}

	b, _ := json.Marshal(createBody)
	createRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/slm/webservice/v2.0/changeset/create", s.cfg.RallyURL), bytes.NewBuffer(b))
	s.DecorateRequest(createRequest)

	createResponse, err := s.client.Do(createRequest)

	if err != nil {
		return err
	}
	defer createResponse.Body.Close()
	var rallyCreateResponse RallyCreateResult
	if err = json.NewDecoder(createResponse.Body).Decode(&rallyCreateResponse); err != nil {
		return err
	}

	changeSetRef := rallyCreateResponse.CreateResult.Object.Ref

	if changeSetRef == "" {
		return errors.New("unable to create changeset")
	}

	// Add changes from commit to changeset
	// For each added, modifed, removed create a change
	for _, a := range c.Added {
		if err := s.AddChange("A", changeSetRef, a, fmt.Sprintf("%s/blob/%s/%s", repoURL, branch, a)); err != nil {
			fmt.Printf("Error adding change: %s", err.Error())
		}
	}

	for _, m := range c.Modified {
		if err := s.AddChange("M", changeSetRef, m, fmt.Sprintf("%s/blob/%s/%s", repoURL, branch, m)); err != nil {
			fmt.Printf("Error adding change: %s", err.Error())
		}
	}

	for _, r := range c.Removed {
		if err := s.AddChange("R", changeSetRef, r, fmt.Sprintf("%s/blob/%s/%s", repoURL, branch, r)); err != nil {
			fmt.Printf("Error adding change: %s", err.Error())
		}
	}
	return err
}

// UpdateState - updates schedulestate in rally
func (s *service) UpdateState(ref string, state string) (err error) {

	updatePayload := map[string]interface{}{
		"HierarchicalRequirement": map[string]interface{}{
			"ScheduleState": state,
		},
	}

	b, _ := json.Marshal(updatePayload)
	updateRequest, _ := http.NewRequest(http.MethodPost, ref, bytes.NewBuffer(b))
	s.DecorateRequest(updateRequest)

	updateResponse, err := s.client.Do(updateRequest)

	if err != nil {
		return
	}
	defer updateResponse.Body.Close()
	var updateResult UpdateResult

	if err = json.NewDecoder(updateResponse.Body).Decode(&updateResult); err != nil {
		return err
	}

	if updateResult.OperationResult.Object.ScheduleState != state {
		return fmt.Errorf("failed to update state - %s", updateResult.OperationResult.Errors)
	}

	return
}

func (s *service) AddChange(action string, changeset string, path string, uri string) error {
	var err error

	createBody := map[string]interface{}{
		"Change": map[string]interface{}{
			"Action":          action,
			"Changeset":       changeset,
			"PathAndFilename": path,
			"Uri":             uri,
		},
	}

	b, _ := json.Marshal(createBody)
	createRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/slm/webservice/v2.0/change/create", s.cfg.RallyURL), bytes.NewBuffer(b))
	s.DecorateRequest(createRequest)

	createResponse, err := s.client.Do(createRequest)

	if err != nil {
		return err
	}
	defer createResponse.Body.Close()
	var rallyCreateResponse RallyCreateResult
	if err = json.NewDecoder(createResponse.Body).Decode(&rallyCreateResponse); err != nil {
		return err
	}

	return err
}

func (s *service) ValidateOrg(orgname string) (string, bool) {
	urlString := fmt.Sprintf("%s/slm/webservice/v2.0/workspace", s.cfg.RallyURL)
	req, _ := http.NewRequest(http.MethodGet, urlString, nil)

	params := url.Values{}
	params.Set("query", fmt.Sprintf("(name = %s)", orgname))

	req.URL.RawQuery = params.Encode()

	s.DecorateRequest(req)

	var rallyresponse RallyQueryResults
	response, err := s.client.Do(req)

	if err != nil {
		return "", false
	}
	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(&rallyresponse); err != nil {
		return "", false
	}

	results := rallyresponse.QueryResult.Results
	if len(results) == 1 {
		return results[0].Ref, true
	}

	return "", false
}

func (s *service) GetOrCreateSCMRepository(repo string, repoURL string, workspace string) (string, error) {
	urlString := fmt.Sprintf("%s/slm/webservice/v2.0/scmrepository", s.cfg.RallyURL)
	req, _ := http.NewRequest(http.MethodGet, urlString, nil)

	params := url.Values{}
	params.Set("query", fmt.Sprintf("(name = %s)", repo))

	req.URL.RawQuery = params.Encode()

	s.DecorateRequest(req)

	var rallyresponse RallyQueryResults
	response, err := s.client.Do(req)

	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(&rallyresponse); err != nil {
		return "", err
	}

	results := rallyresponse.QueryResult.Results
	if len(results) == 1 {
		return results[0].Ref, nil
	}

	createBody := map[string]interface{}{
		"SCMRepository": map[string]interface{}{
			"SCMType":     "GitHub",
			"Name":        repo,
			"Workspace":   workspace,
			"Description": "GitHub-Service push Changesets",
			"Uri":         repoURL,
		},
	}

	b, _ := json.Marshal(createBody)
	createRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/create", urlString), bytes.NewBuffer(b))
	s.DecorateRequest(createRequest)

	createResponse, err := s.client.Do(createRequest)

	if err != nil {
		return "", err
	}
	defer createResponse.Body.Close()
	var rallyCreateResponse RallyCreateResult
	if err = json.NewDecoder(createResponse.Body).Decode(&rallyCreateResponse); err != nil {
		return "", err
	}

	return rallyCreateResponse.CreateResult.Object.Ref, nil
}

func (s *service) DecorateRequest(req *http.Request) {
	req.Header.Set("ZSESSIONID", s.cfg.APIToken)
}

// CheckForStatus - function takes a string as an argument and returns booleans for start and complete if the keywords are found
func (s *service) checkForStatus(message string, artifactID string) (start bool, complete bool) {
	start = false
	complete = false

	var (
		startRegexString     = `(STARTS|BEGINS)\s` + artifactID
		completesRegexString = `(COMPLETES|FINISHES)\s` + artifactID
	)

	startRegex := regexp.MustCompile(startRegexString)
	completeRegex := regexp.MustCompile(completesRegexString)

	if startRegex.MatchString(message) {
		start = true
	}

	if completeRegex.MatchString(message) {
		complete = true
	}

	return
}

func (s *service) FindRallyArtifact(commit Commit) (artifacts map[string]string) {
	typeMap := map[string]string{
		"D":  "defect",
		"DE": "defect",
		"DS": "defectsuite",
		"TA": "task",
		"TC": "testcase",
		"S":  "hierarchicalrequirement",
		"US": "hierarchicalrequirement",
	}
	var (
		artifactRegexString = `(D|DE|DS|TA|TC|S|US)\d+`
		artifactID          string
		artifactType        string
	)
	artifactRegex := regexp.MustCompile(artifactRegexString)

	if artifactRegex.MatchString(commit.Message) {
		result_slice := artifactRegex.FindAllStringSubmatch(commit.Message, -1)

		if len(result_slice) > 0 {
			artifacts = make(map[string]string, len(result_slice))
			for _, v := range result_slice {
				if len(v) > 0 {
					artifactID = v[0]
					artifactType = v[1]

					urlString := fmt.Sprintf("%s/slm/webservice/v2.0/%s", s.cfg.RallyURL, typeMap[artifactType])
					req, _ := http.NewRequest(http.MethodGet, urlString, nil)

					params := url.Values{}
					params.Set("query", fmt.Sprintf("(FormattedID = %s)", artifactID))

					req.URL.RawQuery = params.Encode()

					s.DecorateRequest(req)

					var rallyresponse RallyQueryResults
					response, err := s.client.Do(req)
					if err != nil {
						continue
					}
					defer response.Body.Close()
					if err = json.NewDecoder(response.Body).Decode(&rallyresponse); err != nil {
						continue
					}

					if rallyresponse.QueryResult.TotalResultCount == 0 {
						continue
					}

					artifacts[artifactID] = rallyresponse.QueryResult.Results[0].Ref
				}
			}
		}

	}

	return artifacts

}
