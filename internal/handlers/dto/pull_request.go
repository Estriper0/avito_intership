package dto

import "time"

type PrCreateRequest struct {
	PrId     string `json:"pull_request_id" validate:"required,max=30"`
	PrName   string `json:"pull_request_name" validate:"required,max=200"`
	AuthorId string `json:"author_id" validate:"required,max=30"`
}

type PullRequest struct {
	PrId              string   `json:"pull_request_id"`
	PrName            string   `json:"pull_request_name"`
	AuthorId          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type MergeRequest struct {
	PrId string `json:"pull_request_id" validate:"required,max=30"`
}

type MergeResponse struct {
	PrId              string    `json:"pull_request_id"`
	PrName            string    `json:"pull_request_name"`
	AuthorId          string    `json:"author_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	MergedAt          time.Time `json:"mergedAt"`
}

type ReassignRequest struct {
	PrId          string `json:"pull_request_id" validate:"required,max=30"`
	OldReviewerId string `json:"old_reviewer_id" validate:"required,max=30"`
}

type ReassignResponse struct {
	PR            *PullRequest `json:"pr"`
	NewReviewerId string       `json:"replaced_by"`
}

type MassReassignResponse struct {
	PrId          string `json:"pull_request_id"`
	OldReviewerId string `json:"old_reviewer_id"`
	NewReviewerId string `json:"new_reviewer_id"`
}
