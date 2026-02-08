package gitdtos

import "fmt"

type BLRootErrorPayload struct {
	CriticalErrors      []interface{}      `json:"critical,omitempty"`
	WorkspaceFetchError string             `json:"workspace_fetch_error,omitempty"`
	WorkspaceErrors     []BLWorkspaceError `json:"workspace_errors,omitempty"`
}

func (e *BLRootErrorPayload) Error() string {
	return fmt.Sprintf("critical errors: %v, workspace fetch error: %s, workspace errors: %v", e.CriticalErrors, e.WorkspaceFetchError, e.WorkspaceErrors)
}

func (e *BLRootErrorPayload) IsEmpty() bool {
	return len(e.CriticalErrors) == 0 && e.WorkspaceFetchError == "" && len(e.WorkspaceErrors) == 0
}

type BLWorkspaceError struct {
	WorkspaceSlug            string        `json:"workspace_slug"`
	WorkspaceProcessingError string        `json:"workspace_processing_error,omitempty"`
	RepoFetchError           string        `json:"repo_fetch_error,omitempty"`
	RepoErrors               []BLRepoError `json:"repo_errors,omitempty"`
}

func (e BLWorkspaceError) Error() string {
	return fmt.Sprintf("workspace %s processing error: %s, repo fetch error: %s, repo errors: %v", e.WorkspaceSlug, e.WorkspaceProcessingError, e.RepoFetchError, e.RepoErrors)
}

func (e BLWorkspaceError) IsEmpty() bool {
	return e.WorkspaceProcessingError == "" && e.RepoFetchError == "" && len(e.RepoErrors) == 0
}

type BLRepoError struct {
	RepoID              string          `json:"repo_id"`
	RepoProcessingError string          `json:"repo_processing_error,omitempty"`
	PrFetchError        string          `json:"pr_fetch_error,omitempty"`
	PrErrors            []BLPrError     `json:"pr_errors,omitempty"`
	CommitFetchError    string          `json:"commit_fetch_error,omitempty"`
	CommitErrors        []BLCommitError `json:"commit_errors,omitempty"`
}

func (e BLRepoError) Error() string {
	return fmt.Sprintf("repo %s processing error: %s, pr fetch error: %s, pr errors: %v, commit fetch error: %s, commit errors: %v", e.RepoID, e.RepoProcessingError, e.PrFetchError, e.PrErrors, e.CommitFetchError, e.CommitErrors)
}

func (e BLRepoError) IsEmpty() bool {
	return e.RepoProcessingError == "" && e.PrFetchError == "" && e.CommitFetchError == "" && len(e.PrErrors) == 0 && len(e.CommitErrors) == 0
}

type BLPrError struct {
	PrID              int             `json:"pr_id"`
	PrProcessingError string          `json:"pr_processing_error,omitempty"`
	CommitFetchError  string          `json:"commit_fetch_error,omitempty"`
	CommitErrors      []BLCommitError `json:"commit_errors,omitempty"`
}

func (e BLPrError) Error() string {
	return fmt.Sprintf("pr %d processing error: %s, commit fetch error: %s, commit errors: %v", e.PrID, e.PrProcessingError, e.CommitFetchError, e.CommitErrors)
}

func (e BLPrError) IsEmpty() bool {
	return e.PrProcessingError == "" && e.CommitFetchError == "" && len(e.CommitErrors) == 0
}

type BLCommitError struct {
	CommitID              string               `json:"commit_id"`
	CommitProcessingError string               `json:"commit_processing_error,omitempty"`
	ChangedFileErrors     []BLChangedFileError `json:"changed_file_errors,omitempty"`
}

func (e BLCommitError) Error() string {
	return fmt.Sprintf("commit %s processing error: %s, changed file errors: %v", e.CommitID, e.CommitProcessingError, e.ChangedFileErrors)
}

func (e BLCommitError) IsEmpty() bool {
	return e.CommitProcessingError == "" && len(e.ChangedFileErrors) == 0
}

type BLChangedFileError struct {
	Filename                   string `json:"filename"`
	ChangedFileProcessingError string `json:"changed_file_processing_error,omitempty"`
}

func (e BLChangedFileError) Error() string {
	return fmt.Sprintf("changed file %s processing error: %s", e.Filename, e.ChangedFileProcessingError)
}

func (e BLChangedFileError) IsEmpty() bool {
	return e.ChangedFileProcessingError == ""
}

type ErrorWithIsEmpty interface {
	error
	IsEmpty() bool
}

var _ ErrorWithIsEmpty = (*BLRootErrorPayload)(nil)
var _ ErrorWithIsEmpty = (*BLWorkspaceError)(nil)
var _ ErrorWithIsEmpty = (*BLRepoError)(nil)
var _ ErrorWithIsEmpty = (*BLPrError)(nil)
var _ ErrorWithIsEmpty = (*BLCommitError)(nil)
var _ ErrorWithIsEmpty = (*BLChangedFileError)(nil)
