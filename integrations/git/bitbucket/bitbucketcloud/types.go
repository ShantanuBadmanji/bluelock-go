package bitbucketcloud

import "time"

type BBktCloudPaginatedResponse[T any] struct {
	Values []T    `json:"values"`
	Next   string `json:"next"`
}

type BBktCloudWorkspace struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type BBktCloudRepository struct {
	Slug      string         `json:"slug"`
	Name      string         `json:"name"`
	ID        string         `json:"uuid"`
	IsPrivate bool           `json:"is_private"`
	Links     BBktCloudLinks `json:"links"`
}

type BBktCloudLinks struct {
	HTML BBktCloudLink `json:"html"`
}
type BBktCloudLink struct {
	Href string `json:"href"`
}

type BBktCloudPullRequest struct {
	ID           int                `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	State        string             `json:"state"`
	CreatedOn    time.Time          `json:"created_on"`
	UpdatedOn    time.Time          `json:"updated_on"`
	ClosedOn     time.Time          `json:"closed_on"`
	Author       BBKtCloudUser      `json:"author"`
	Reviewers    []BBKtCloudUser    `json:"reviewers"`
	Source       BBktCloudBranchRef `json:"source"`
	Destination  BBktCloudBranchRef `json:"destination"`
	CommentCount int                `json:"comment_count"`
	Links        BBktCloudLinks     `json:"links"`
	Draft        bool               `json:"draft"`
}

type BBktCloudPullRequestState string

const (
	BBktCloudPullRequestStateOpen       BBktCloudPullRequestState = "OPEN"
	BBktCloudPullRequestStateMerged     BBktCloudPullRequestState = "MERGED"
	BBktCloudPullRequestStateDeclined   BBktCloudPullRequestState = "DECLINED"
	BBktCloudPullRequestStateSuperseded BBktCloudPullRequestState = "SUPERSEDED"
)

type BBktCloudBranchRef struct {
	Branch     BBktCloudBranch     `json:"branch"`
	Commit     BBktCloudCommit     `json:"commit"`
	Repository BBktCloudRepository `json:"repository"`
}

type BBktCloudBranch struct {
	Name string `json:"name"`
}

type BBCloudSlimCommit struct {
	Hash  string         `json:"hash"`
	Links BBktCloudLinks `json:"links"`
}
type BBktCloudCommit struct {
	Hash    string              `json:"hash"`
	Message string              `json:"message"`
	Author  BBKtCloudActor      `json:"author"`
	Date    time.Time           `json:"date"`
	Links   BBktCloudLinks      `json:"links"`
	Parents []BBCloudSlimCommit `json:"parents"`
}

type BBKtCloudActor struct {
	Raw  string        `json:"raw"`
	User BBKtCloudUser `json:"user"`
}

type BBKtCloudUser struct {
	UUID        string         `json:"uuid"`
	DisplayName string         `json:"display_name"`
	AccountID   string         `json:"account_id"`
	Links       BBktCloudLinks `json:"links"`
}
