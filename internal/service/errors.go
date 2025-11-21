package service

import "errors"

var (
	ErrTeamAlreadyExists        = errors.New("team_name already exists")
	ErrPullRequestALreadyExists = errors.New("pr id already exists")

	ErrPullRequestMerged = errors.New("cannot reassign on merged PR")
	ErrNoCandidate       = errors.New("no candidate for reassign")

	ErrNotFound = errors.New("resource not found")
	ErrInternal = errors.New("internal error")
)
