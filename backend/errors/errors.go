package custom_errors

import "errors"

var (
	ErrDuplicateData    error = errors.New("data already exists")
	ErrNoRecordFound    error = errors.New("no record found")
	ErrResourceConflict error = errors.New("resource tampered")
	ErrUploadIssue      error = errors.New("upload issue, either it is pending or failed")
	ErrResourceLocked   error = errors.New("resource locked please try later")
)
