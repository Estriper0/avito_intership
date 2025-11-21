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
