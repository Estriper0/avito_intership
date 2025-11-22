package dto

type Team struct {
	TeamName string    `json:"team_name" validate:"required,max=30"`
	Members  []Members `json:"members" validate:"required,min=1"`
}

type Members struct {
	UserId   string `json:"user_id" validate:"required,max=30"`
	Username string `json:"username" validate:"required,max=30"`
	IsActive bool   `json:"is_active" validate:"required"`
}

type TeamStatsPrResponse struct {
	Name     string `json:"team_name"`
	TotalPr  int    `json:"total_pull_request"`
	OpenPr   int    `json:"open_pull_request"`
	MergedPr int    `json:"merged_pull_request"`
}
