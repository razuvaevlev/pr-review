package repo

import "pr-review/internal/entity"

type PullRequestRepository interface {
	CreatePR(pr *entity.PullRequest) error

	GetPR(prID string) (*entity.PullRequest, error)

	UpdatePR(pr *entity.PullRequest) error

	PRExists(prID string) (bool, error)

	GetPRsByReviewer(userID string) ([]*entity.PullRequest, error)
}
