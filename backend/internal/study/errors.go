package study

import "errors"

var (
	// Sessions
	ErrNotFound       = errors.New("study session not found")
	ErrMissingSubject = errors.New("subject is required")
	ErrInvalidTiming  = errors.New("start and end time must be provided and end must be after start")
	ErrUnknownSubject = errors.New("subject does not exist")

	// Subjects
	ErrSubjectNotFound   = errors.New("subject not found")
	ErrSubjectNameExists = errors.New("subject name already exists")
	ErrSubjectNameEmpty  = errors.New("subject name is required")
)
