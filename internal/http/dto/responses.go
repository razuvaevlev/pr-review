package dto

import "pr-review/internal/entity"

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type TeamResponse struct {
	Team *entity.Team `json:"team,omitempty"`
}

type UserResponse struct {
	User *entity.User `json:"user,omitempty"`
}

type GetReviewResponse struct {
	UserID       string                     `json:"user_id"`
	PullRequests []*entity.PullRequestShort `json:"pull_requests"`
}
