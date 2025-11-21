package models

type User struct {
	UserId   string
	Username string
	TeamId   int
	IsActive bool
}

type UserStatsReview struct {
	UserId          string
	Username        string
	CountOpenReview int
}
