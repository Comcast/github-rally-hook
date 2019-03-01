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

import "time"

type Config struct {
	RallyURL          string    `json:"rally-url"`
	APIToken          string    `json:"api-key"`
	Workspace         string    `json:"workspace"`
	SecretToken       string    `json:"secret_token"`
	SignatureRequired bool      `json:"signature_required"`
	InfluxCfg         InfluxCfg `json:"influx_cfg"`
}

// InfluxCfg - struct
type InfluxCfg struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Tag      string `json:"tag"`
}

type PushResponse struct {
	Result string  `json:"result"`
	Errors []error `json:"errors"`
}

type Reference struct {
	Count         int    `json:",omitempty"`
	Ref           string `json:"_ref,omitempty"`
	Type          string `json:"_type,omitempty"`
	RefObjectName string `json:"_refObjectName,omitempty"`
	RefObjectUUID string `json:"_refObjectUUID,omitempty"`
}

type Changeset struct {
	Ref             string      `json:"_ref,omitempty"`
	CreationDate    string      `json:",omitempty"`
	ObjectID        int         `json:",omitempty"`
	ObjectUUID      string      `json:",omitempty"`
	Subscription    string      `json:",omitempty"`
	Workspace       string      `json:",omitempty"`
	Artifacts       []Reference `json:",omitempty"`
	Author          string      `json:",omitempty"`
	Branch          string      `json:",omitempty"`
	Builds          string      `json:",omitempty"`
	Changes         string      `json:",omitempty"`
	CommitTimestamp string      `json:",omitempty"`
	Message         string      `json:",omitempty"`
	Name            string      `json:",omitempty"`
	Revision        string      `json:",omitempty"`
	SCMRepository   string      `json:",omitempty"`
	Uri             string      `json:",omitempty"`
}

type RallyQueryResults struct {
	QueryResult struct {
		RallyAPIMajor    string        `json:"_rallyAPIMajor"`
		RallyAPIMinor    string        `json:"_rallyAPIMinor"`
		Errors           []interface{} `json:"Errors"`
		Warnings         []interface{} `json:"Warnings"`
		TotalResultCount float64       `json:"TotalResultCount"`
		StartIndex       int           `json:"StartIndex"`
		PageSize         int           `json:"PageSize"`
		Results          []RallyResult `json:"Results"`
	} `json:"QueryResult"`
}

type RallyResult struct {
	RallyAPIMajor string `json:"_rallyAPIMajor"`
	RallyAPIMinor string `json:"_rallyAPIMinor"`
	Ref           string `json:"_ref"`
	RefObjectUUID string `json:"_refObjectUUID"`
	RefObjectName string `json:"_refObjectName"`
	Type          string `json:"_type"`
}

type Commit struct {
	ID        string `json:"id"`
	TreeID    string `json:"tree_id"`
	Distinct  bool   `json:"distinct"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
	Committer struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"committer"`
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

type PushEvent struct {
	Ref        string      `json:"ref"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
	Created    bool        `json:"created"`
	Deleted    bool        `json:"deleted"`
	Forced     bool        `json:"forced"`
	BaseRef    interface{} `json:"base_ref"`
	Compare    string      `json:"compare"`
	Commits    []Commit    `json:"commits"`
	HeadCommit interface{} `json:"head_commit"`
	Repository struct {
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Owner    struct {
			Name              string `json:"name"`
			Email             string `json:"email"`
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"owner"`
		Private          bool        `json:"private"`
		HTMLURL          string      `json:"html_url"`
		Description      interface{} `json:"description"`
		Fork             bool        `json:"fork"`
		URL              string      `json:"url"`
		ForksURL         string      `json:"forks_url"`
		KeysURL          string      `json:"keys_url"`
		CollaboratorsURL string      `json:"collaborators_url"`
		TeamsURL         string      `json:"teams_url"`
		HooksURL         string      `json:"hooks_url"`
		IssueEventsURL   string      `json:"issue_events_url"`
		EventsURL        string      `json:"events_url"`
		AssigneesURL     string      `json:"assignees_url"`
		BranchesURL      string      `json:"branches_url"`
		TagsURL          string      `json:"tags_url"`
		BlobsURL         string      `json:"blobs_url"`
		GitTagsURL       string      `json:"git_tags_url"`
		GitRefsURL       string      `json:"git_refs_url"`
		TreesURL         string      `json:"trees_url"`
		StatusesURL      string      `json:"statuses_url"`
		LanguagesURL     string      `json:"languages_url"`
		StargazersURL    string      `json:"stargazers_url"`
		ContributorsURL  string      `json:"contributors_url"`
		SubscribersURL   string      `json:"subscribers_url"`
		SubscriptionURL  string      `json:"subscription_url"`
		CommitsURL       string      `json:"commits_url"`
		GitCommitsURL    string      `json:"git_commits_url"`
		CommentsURL      string      `json:"comments_url"`
		IssueCommentURL  string      `json:"issue_comment_url"`
		ContentsURL      string      `json:"contents_url"`
		CompareURL       string      `json:"compare_url"`
		MergesURL        string      `json:"merges_url"`
		ArchiveURL       string      `json:"archive_url"`
		DownloadsURL     string      `json:"downloads_url"`
		IssuesURL        string      `json:"issues_url"`
		PullsURL         string      `json:"pulls_url"`
		MilestonesURL    string      `json:"milestones_url"`
		NotificationsURL string      `json:"notifications_url"`
		LabelsURL        string      `json:"labels_url"`
		ReleasesURL      string      `json:"releases_url"`
		DeploymentsURL   string      `json:"deployments_url"`
		CreatedAt        int         `json:"created_at"`
		UpdatedAt        time.Time   `json:"updated_at"`
		PushedAt         int         `json:"pushed_at"`
		GitURL           string      `json:"git_url"`
		SSHURL           string      `json:"ssh_url"`
		CloneURL         string      `json:"clone_url"`
		SvnURL           string      `json:"svn_url"`
		Homepage         interface{} `json:"homepage"`
		Size             int         `json:"size"`
		StargazersCount  int         `json:"stargazers_count"`
		WatchersCount    int         `json:"watchers_count"`
		Language         interface{} `json:"language"`
		HasIssues        bool        `json:"has_issues"`
		HasProjects      bool        `json:"has_projects"`
		HasDownloads     bool        `json:"has_downloads"`
		HasWiki          bool        `json:"has_wiki"`
		HasPages         bool        `json:"has_pages"`
		ForksCount       int         `json:"forks_count"`
		MirrorURL        interface{} `json:"mirror_url"`
		Archived         bool        `json:"archived"`
		OpenIssuesCount  int         `json:"open_issues_count"`
		License          interface{} `json:"license"`
		Forks            int         `json:"forks"`
		OpenIssues       int         `json:"open_issues"`
		Watchers         int         `json:"watchers"`
		DefaultBranch    string      `json:"default_branch"`
		Stargazers       int         `json:"stargazers"`
		MasterBranch     string      `json:"master_branch"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Sender struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"sender"`
}

type RallyCreateResult struct {
	CreateResult struct {
		RallyAPIMajor string        `json:"_rallyAPIMajor"`
		RallyAPIMinor string        `json:"_rallyAPIMinor"`
		Errors        []interface{} `json:"Errors"`
		Warnings      []interface{} `json:"Warnings"`
		Object        struct {
			RallyAPIMajor string    `json:"_rallyAPIMajor"`
			RallyAPIMinor string    `json:"_rallyAPIMinor"`
			Ref           string    `json:"_ref"`
			RefObjectUUID string    `json:"_refObjectUUID"`
			ObjectVersion string    `json:"_objectVersion"`
			RefObjectName string    `json:"_refObjectName"`
			CreationDate  time.Time `json:"CreationDate"`
			CreatedAt     string    `json:"_CreatedAt"`
			ObjectID      int64     `json:"ObjectID"`
			ObjectUUID    string    `json:"ObjectUUID"`
			VersionID     string    `json:"VersionId"`
			Subscription  struct {
				RallyAPIMajor string `json:"_rallyAPIMajor"`
				RallyAPIMinor string `json:"_rallyAPIMinor"`
				Ref           string `json:"_ref"`
				RefObjectUUID string `json:"_refObjectUUID"`
				RefObjectName string `json:"_refObjectName"`
				Type          string `json:"_type"`
			} `json:"Subscription"`
			Workspace struct {
				RallyAPIMajor string `json:"_rallyAPIMajor"`
				RallyAPIMinor string `json:"_rallyAPIMinor"`
				Ref           string `json:"_ref"`
				RefObjectUUID string `json:"_refObjectUUID"`
				RefObjectName string `json:"_refObjectName"`
				Type          string `json:"_type"`
			} `json:"Workspace"`
			Description string `json:"Description"`
			Name        string `json:"Name"`
			Projects    struct {
				RallyAPIMajor string `json:"_rallyAPIMajor"`
				RallyAPIMinor string `json:"_rallyAPIMinor"`
				Ref           string `json:"_ref"`
				Type          string `json:"_type"`
				Count         int    `json:"Count"`
			} `json:"Projects"`
			SCMType string `json:"SCMType"`
			URI     string `json:"Uri"`
			Type    string `json:"_type"`
		} `json:"Object"`
	} `json:"CreateResult"`
}
