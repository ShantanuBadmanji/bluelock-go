package git

import "time"

type DevDRootPayload[T interface{}] struct {
	Data []T `json:"data,omitempty"`
}

type DevDChangedFile struct {
	Filename   string `json:"filename"`
	ChangeType string `json:"change_type"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	NewWork    int    `json:"new_work"`
	Refactor   int    `json:"refactor"`
	Rework     int    `json:"rework"`
	HelpOthers int    `json:"help_others"`
}

type DevDActor struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}
type DevDCommit struct {
	ID                 string            `json:"id"`
	Message            string            `json:"message"`
	Committer          DevDActor         `json:"committer"`
	CommitterTimestamp time.Time         `json:"committerTimestamp"`
	ChangedFiles       []DevDChangedFile `json:"changed_files"`
}

type DevDAdditionalParam struct {
	Reviewer interface{} `json:"reviewer"`
}

type DevDAdditionalParam1 struct {
	DevDAdditionalParam
	Reviewer interface{} `json:"reviewer"`
}

type DevDActivityInfo struct {
	ID               string               `json:"id"`
	Type             string               `json:"type"`
	Action           string               `json:"action"`
	Actor            DevDActor            `json:"actor"`
	UpdatedAt        float64              `json:"updated_at"`
	AdditionalParam1 DevDAdditionalParam1 `json:"additional_param,omitempty"`
}

type DevDPullRequest struct {
	CommitsPr    []DevDCommit       `json:"commits_pr"`
	ActivityInfo []DevDActivityInfo `json:"activity_info"`
	ID           int                `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	State        string             `json:"state"`
	Open         bool               `json:"open"`
	Closed       bool               `json:"closed"`
	CreatedDate  float64            `json:"createdDate"`
	UpdatedDate  float64            `json:"updatedDate"`
	SourceBranch string             `json:"sourceBranch"`
	TargetBranch string             `json:"targetBranch"`
	Author       DevDActor          `json:"author"`
	Reviewers    []interface{}      `json:"reviewers"`
	CommentCount int                `json:"commentCount"`
	Link         string             `json:"link"`
}

type DevDRepo struct {
	Prs      []DevDPullRequest `json:"prs"`
	Slug     string            `json:"slug"`
	Name     string            `json:"name"`
	ID       string            `json:"id"`
	IsPublic bool              `json:"isPublic"`
	Link     string            `json:"link"`
	Commits  []DevDCommit      `json:"commits"`
}

type DevDData struct {
	Repos        []DevDRepo `json:"repos"`
	WorkspaceKey string     `json:"workspaceKey"`
}

type DevDError struct {
	Critical         []interface{} `json:"critical,omitempty"`
	ProjectsErrorLog interface{}   `json:"projects_error_log,omitempty"`
}
