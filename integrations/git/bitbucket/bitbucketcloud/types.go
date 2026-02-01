package bitbucketcloud

type BBktCloudPaginatedResponse[T any] struct {
	Values []T    `json:"values"`
	Next   string `json:"next"`
}

type BBktCloudWorkspace struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type BBktCloudRepository struct {
	Slug      string           `json:"slug"`
	Name      string           `json:"name"`
	ID        string           `json:"uuid"`
	IsPrivate bool             `json:"is_private"`
	Links     BBktCloudLinks `json:"links"`
}

type BBktCloudLinks struct {
	HTML BBktCloudLink `json:"html"`
}
type BBktCloudLink struct {
	Href string `json:"href"`
}

type BBktCloudPullRequest struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	State string `json:"state"`
}

type BBktCloudCommit struct {
	Hash    string         `json:"hash"`
	Message string         `json:"message"`
	Author  BBKtCloudActor `json:"author"`
}

type BBKtCloudActor struct {
	Raw string `json:"raw"`
}
