package models

import "time"

type PullRequest struct {
	PrId      string
	Name      string
	AuthorId  string
	StatusId  int
	CreatedAt time.Time
	MergedAt  time.Time
}
