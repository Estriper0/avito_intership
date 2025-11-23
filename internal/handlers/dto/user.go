package dto

type SetIsActiveRequest struct {
	UserId   string `json:"user_id" validate:"required,max=30"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

type UserResponse struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type ReviewResponse struct {
	PrId     string `json:"pull_request_id"`
	PrName   string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
	Status   string `json:"status"`
}

type UserStatsReviewResponse struct {
	UserId          string `json:"user_id"`
	Username        string `json:"username"`
	CountOpenReview int    `json:"count_open_review"`
}

type MassDeactivationRequest struct {
	UsersId []string `json:"users_id" validate:"min=1"`
}

type MassDeactivationResponse struct {
	UsersId []string `json:"deactivated_users_id"`
}
