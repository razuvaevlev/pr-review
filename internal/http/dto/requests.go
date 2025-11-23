package dto

import (
	"errors"
	"pr-review/internal/config"
	"strings"
)

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func (r *SetIsActiveRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("user_id cannot be empty")
	}
	if len(r.UserID) > config.MaxStringLength {
		return errors.New("user_id cannot exceed 255 characters")
	}
	return nil
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

func (r *CreatePRRequest) Validate() error {
	if strings.TrimSpace(r.PullRequestID) == "" {
		return errors.New("pull_request_id cannot be empty")
	}
	if len(r.PullRequestID) > config.MaxStringLength {
		return errors.New("pull_request_id cannot exceed 255 characters")
	}
	if strings.TrimSpace(r.PullRequestName) == "" {
		return errors.New("pull_request_name cannot be empty")
	}
	if len(r.PullRequestName) > config.MaxStringLength {
		return errors.New("pull_request_name cannot exceed 255 characters")
	}
	if strings.TrimSpace(r.AuthorID) == "" {
		return errors.New("author_id cannot be empty")
	}
	if len(r.AuthorID) > config.MaxStringLength {
		return errors.New("author_id cannot exceed 255 characters")
	}
	return nil
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

func (r *MergePRRequest) Validate() error {
	if strings.TrimSpace(r.PullRequestID) == "" {
		return errors.New("pull_request_id cannot be empty")
	}
	if len(r.PullRequestID) > config.MaxStringLength {
		return errors.New("pull_request_id cannot exceed 255 characters")
	}
	return nil
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_user_id" binding:"required"`
}

func (r *ReassignReviewerRequest) Validate() error {
	if strings.TrimSpace(r.PullRequestID) == "" {
		return errors.New("pull_request_id cannot be empty")
	}
	if len(r.PullRequestID) > config.MaxStringLength {
		return errors.New("pull_request_id cannot exceed 255 characters")
	}
	if strings.TrimSpace(r.OldUserID) == "" {
		return errors.New("old_user_id cannot be empty")
	}
	if len(r.OldUserID) > config.MaxStringLength {
		return errors.New("old_user_id cannot exceed 255 characters")
	}
	return nil
}
