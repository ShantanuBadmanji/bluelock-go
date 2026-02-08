package gitdtos

import "time"

type BLRootPayload[T interface{}] struct {
	Data []T `json:"data,omitempty"`
}

type BLChangedFile struct {
	Filename   string `json:"filename"`
	ChangeType string `json:"change_type"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	NewWork    int    `json:"new_work"`
	Refactor   int    `json:"refactor"`
	Rework     int    `json:"rework"`
	HelpOthers int    `json:"help_others"`
}

type BLActor struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}
type BLCommit struct {
	ID                 string          `json:"id"`
	Message            string          `json:"message"`
	Committer          BLActor         `json:"committer"`
	CommitterTimestamp time.Time       `json:"committerTimestamp"`
	ChangedFiles       []BLChangedFile `json:"changed_files"`
}

type BLAdditionalParam struct {
	Reviewer interface{} `json:"reviewer"`
}

type BLAdditionalParam1 struct {
	BLAdditionalParam
	Reviewer interface{} `json:"reviewer"`
}

type BLActivityInfo struct {
	ID               string             `json:"id"`
	Type             string             `json:"type"`
	Action           string             `json:"action"`
	Actor            BLActor            `json:"actor"`
	UpdatedAt        time.Time          `json:"updated_at"`
	AdditionalParam1 BLAdditionalParam1 `json:"additional_param,omitempty"`
}

type BLPullRequest struct {
	PrCommits    []BLCommit       `json:"pr_commits"`
	ActivityInfo []BLActivityInfo `json:"activity_info"`
	ID           int              `json:"id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	State        string           `json:"state"`
	Open         bool             `json:"open"`
	Closed       bool             `json:"closed"`
	CreatedDate  time.Time        `json:"createdDate"`
	UpdatedDate  time.Time        `json:"updatedDate"`
	SourceBranch string           `json:"sourceBranch"`
	TargetBranch string           `json:"targetBranch"`
	Author       BLActor          `json:"author"`
	Reviewers    []BLActor        `json:"reviewers"`
	CommentCount int              `json:"commentCount"`
	Link         string           `json:"link"`
}

func (p BLPullRequest) IsEmpty() bool {
	return len(p.PrCommits) == 0 && len(p.ActivityInfo) == 0
}

type BLRepo struct {
	Slug     string          `json:"slug"`
	Name     string          `json:"name"`
	ID       string          `json:"id"`
	IsPublic bool            `json:"isPublic"`
	Link     string          `json:"link"`
	Commits  []BLCommit      `json:"commits"`
	Prs      []BLPullRequest `json:"prs"`
}

func (r BLRepo) IsEmpty() bool {
	return len(r.Commits) == 0 && len(r.Prs) == 0
}

type BLData struct {
	Repos        []BLRepo `json:"repos"`
	WorkspaceKey string   `json:"workspaceKey"`
}
