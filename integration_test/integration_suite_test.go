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
	"fmt"
	"github.com/comcast/rally-rest-toolkit"
	"github.com/comcast/rally-rest-toolkit/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"os"
	"strconv"

	"testing"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "rally-github-service integration test suite")
}

var (
	APIKey    string
	ProjectID string
	StoryID   string
	err       error
)
var _ = BeforeSuite(func() {
	APIKey = os.Getenv("API_KEY")
	if APIKey == "" {
		Skip("API_KEY must be provided")
	}
	ProjectID = os.Getenv("PROJECT_ID")
	if ProjectID == "" {
		Skip("PROJECT_ID must be provided")
	}

	StoryID, err = CreateRallyStory(APIKey, ProjectID)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	err = DeleteRallyStory(APIKey, StoryID)
	Expect(err).ShouldNot(HaveOccurred())
})

func CreateRallyStory(apiKey string, projectID string) (storyID string, err error) {

	rallyClient := rallyresttoolkit.New(apiKey, "https://rally1.rallydev.com/slm/webservice/v2.0", &http.Client{})
	hrclient := rallyresttoolkit.NewHierarchicalRequirement(rallyClient)

	hrModel := models.HierarchicalRequirement{
		Name: "concourse test story",
		Project: &models.Reference{
			Ref:  fmt.Sprintf("%s/%s", "https://rally1.rallydev.com/slm/webservice/v2.0/project", projectID),
			Type: "Project",
		},
	}

	hr, err := hrclient.CreateHierarchicalRequirement(hrModel)

	if err != nil {
		fmt.Printf("ERROR: %s", err.Error())
	}
	fmt.Printf("Created Story: %s\n", hr.FormattedID)

	return hr.FormattedID, err
}

func DeleteRallyStory(apiKey string, storyID string) (err error) {

	rallyClient := rallyresttoolkit.New(apiKey, "https://rally1.rallydev.com/slm/webservice/v2.0", &http.Client{})
	hrclient := rallyresttoolkit.NewHierarchicalRequirement(rallyClient)

	query := map[string]string{
		"FormattedID": storyID,
	}

	hrs, err := hrclient.QueryHierarchicalRequirement(query)

	if len(hrs) == 1 {
		objectID := strconv.Itoa(hrs[0].ObjectID)
		err = hrclient.DeleteHierarchicalRequirement(objectID)
		if err == nil {
			fmt.Println("Story Deleted")
		}
	}

	return err
}

func GetChangeSetCount(apiKey string, storyID string) (cnt int, err error) {

	rallyClient := rallyresttoolkit.New(apiKey, "https://rally1.rallydev.com/slm/webservice/v2.0", &http.Client{})
	hrclient := rallyresttoolkit.NewHierarchicalRequirement(rallyClient)

	query := map[string]string{
		"FormattedID": storyID,
	}

	hrs, err := hrclient.QueryHierarchicalRequirement(query)

	if len(hrs) == 1 {
		objectID := strconv.Itoa(hrs[0].ObjectID)
		hr, _ := hrclient.GetHierarchicalRequirement(objectID)

		cnt = hr.Changesets.Count
		fmt.Fprintf(os.Stderr, "Found %v changesets\n", cnt)
	}
	return cnt, err
}

func GetStoryScheduledState(apiKey string, storyID string) (state string, err error) {
	rallyClient := rallyresttoolkit.New(apiKey, "https://rally1.rallydev.com/slm/webservice/v2.0", &http.Client{})
	hrclient := rallyresttoolkit.NewHierarchicalRequirement(rallyClient)

	query := map[string]string{
		"FormattedID": storyID,
	}

	hrs, err := hrclient.QueryHierarchicalRequirement(query)

	if len(hrs) == 1 {
		objectID := strconv.Itoa(hrs[0].ObjectID)
		hr, _ := hrclient.GetHierarchicalRequirement(objectID)

		state = hr.ScheduleState
		fmt.Fprintf(os.Stderr, "Found state - %s\n", state)
	}
	return state, err
}
