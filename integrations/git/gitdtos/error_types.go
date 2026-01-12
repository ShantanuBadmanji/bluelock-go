package gitdtos

import "fmt"

type BLRootErrorPayload struct {
	CriticalErrors  []interface{}      `json:"critical,omitempty"`
	WorkspaceErrors []BLWorkspaceError `json:"workspace_errors,omitempty"`
}

func (e *BLRootErrorPayload) Error() string {
	return fmt.Sprintf("critical errors: %v, workspace errors: %v", e.CriticalErrors, e.WorkspaceErrors)
}

func (e *BLRootErrorPayload) IsEmpty() bool {
	return len(e.CriticalErrors) == 0 && len(e.WorkspaceErrors) == 0
}

type BLWorkspaceError struct {
	WorkspaceSlug            string        `json:"workspace_slug"`
	WorkspaceProcessingError string        `json:"workspace_processing_error,omitempty"`
	RepoFetchError           string        `json:"repo_fetch_error,omitempty"`
	RepoErrors               []BLRepoError `json:"repo_errors,omitempty"`
}

type BLRepoError struct {
	RepoID              string          `json:"repo_id"`
	RepoProcessingError string          `json:"repo_processing_error,omitempty"`
	PrFetchError        string          `json:"pr_fetch_error,omitempty"`
	PrErrors            []BLPrError     `json:"pr_errors,omitempty"`
	CommitErrors        []BLCommitError `json:"commit_errors,omitempty"`
}

type BLPrError struct {
	PrID              int             `json:"pr_id"`
	PrProcessingError string          `json:"pr_processing_error,omitempty"`
	CommitFetchError  string          `json:"commit_fetch_error,omitempty"`
	CommitErrors      []BLCommitError `json:"commit_errors,omitempty"`
}

type BLCommitError struct {
	CommitID              string               `json:"commit_id"`
	CommitProcessingError string               `json:"commit_processing_error,omitempty"`
	ChangedFileErrors     []BLChangedFileError `json:"changed_file_errors,omitempty"`
}

type BLChangedFileError struct {
	Filename                   string `json:"filename"`
	ChangedFileProcessingError string `json:"changed_file_processing_error,omitempty"`
}

type EmptyError interface {
	Error() string
	IsEmpty() bool
}

var _ EmptyError = (*BLRootErrorPayload)(nil)
var _ EmptyError = (*BLWorkspaceError)(nil)
var _ EmptyError = (*BLRepoError)(nil)
var _ EmptyError = (*BLPrError)(nil)
var _ EmptyError = (*BLCommitError)(nil)
var _ EmptyError = (*BLChangedFileError)(nil)
