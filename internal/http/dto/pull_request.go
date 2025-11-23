package dto

import (
	"pr-review/internal/entity"
	"time"
)

type PullRequestDTO struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

func FromEntity(pr *entity.PullRequest) *PullRequestDTO {
	if pr == nil {
		return nil
	}

	return &PullRequestDTO{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

type PullRequestResponse struct {
	PR *PullRequestDTO `json:"pr,omitempty"`
}

type ReassignResponse struct {
	PR         *PullRequestDTO `json:"pr"`
	ReplacedBy string          `json:"replaced_by"`
}
