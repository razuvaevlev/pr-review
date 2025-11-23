package service

import (
	crand "crypto/rand"
	"math/big"
	"pr-review/internal/config"
	"pr-review/internal/entity"
	"pr-review/internal/logging"
	"pr-review/internal/repo"
	"time"
)

type PullRequestService struct {
	prRepo   repo.PullRequestRepository
	userRepo repo.UserRepository
	teamRepo repo.TeamRepository
}

func NewPullRequestService(
	prRepo repo.PullRequestRepository,
	userRepo repo.UserRepository,
	teamRepo repo.TeamRepository,
) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *PullRequestService) CreatePR(prID, prName, authorID string) (*entity.PullRequest, error) {
	if derr := s.validateField("pull_request_id", prID); derr != nil {
		return nil, derr
	}
	if derr := s.validateField("pull_request_name", prName); derr != nil {
		return nil, derr
	}
	if derr := s.validateField("author_id", authorID); derr != nil {
		return nil, derr
	}
	exists, err := s.prRepo.PRExists(prID)
	if err != nil {
		logging.Printf("ERROR: Failed to check if PR exists %s: %v", prID, err)
		return nil, err
	}
	if exists {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodePRExists,
			Message: "PR id already exists",
		}
	}

	candidates, err := s.getAuthorAndCandidates(authorID)
	if err != nil {
		return nil, err
	}

	reviewerIDs := s.selectReviewers(candidates, config.DefaultReviewers)
	reviewersStr := reviewerIDs

	now := time.Now()
	pr := &entity.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            entity.StatusOpen,
		AssignedReviewers: reviewersStr,
		CreatedAt:         &now,
		MergedAt:          nil,
	}

	if err := s.prRepo.CreatePR(pr); err != nil {
		logging.Printf("ERROR: Failed to create PR %s: %v", prID, err)
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) MergePR(prID string) (*entity.PullRequest, error) {
	if prID == "" {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "pull_request_id cannot be empty",
		}
	}
	if len(prID) > config.MaxStringLength {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "pull_request_id cannot exceed 255 characters",
		}
	}

	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		logging.Printf("ERROR: Failed to get PR %s: %v", prID, err)
		return nil, err
	}
	if pr == nil {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "PR not found",
		}
	}

	if pr.Status == entity.StatusMerged {
		return pr, nil
	}

	now := time.Now()
	pr.Status = entity.StatusMerged
	pr.MergedAt = &now

	if err := s.prRepo.UpdatePR(pr); err != nil {
		logging.Printf("ERROR: Failed to update PR %s: %v", prID, err)
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(prID, oldUserID string) (*entity.PullRequest, string, error) {
	if derr := s.validateField("pull_request_id", prID); derr != nil {
		return nil, "", derr
	}
	if derr := s.validateField("old_user_id", oldUserID); derr != nil {
		return nil, "", derr
	}

	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		logging.Printf("ERROR: Failed to get PR %s: %v", prID, err)
		return nil, "", err
	}
	if pr == nil {
		return nil, "", &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "PR not found",
		}
	}

	if pr.Status == entity.StatusMerged {
		return nil, "", &entity.DomainError{
			Code:    entity.ErrorCodePRMerged,
			Message: "cannot reassign on merged PR",
		}
	}

	reviewers := pr.AssignedReviewers

	if !s.containsReviewer(reviewers, oldUserID) {
		return nil, "", &entity.DomainError{
			Code:    entity.ErrorCodeNotAssigned,
			Message: "reviewer is not assigned to this PR",
		}
	}
	candidates, err := s.buildReplacementCandidates(pr, oldUserID, reviewers)
	if err != nil {
		return nil, "", err
	}

	if len(candidates) == 0 {
		return nil, "", &entity.DomainError{
			Code:    entity.ErrorCodeNoCandidate,
			Message: "no active replacement candidate in team",
		}
	}

	newUserID := s.selectReviewers(candidates, 1)[0]

	if err := s.applyReplacement(pr, oldUserID, newUserID, reviewers); err != nil {
		return nil, "", err
	}

	return pr, newUserID, nil
}

func (s *PullRequestService) applyReplacement(pr *entity.PullRequest, oldUserID, newUserID string, reviewers []string) error {
	newReviewers := make([]string, 0, len(reviewers))
	for _, reviewer := range reviewers {
		if reviewer != oldUserID {
			newReviewers = append(newReviewers, reviewer)
		}
	}
	newReviewers = append(newReviewers, newUserID)
	pr.AssignedReviewers = newReviewers

	if err := s.prRepo.UpdatePR(pr); err != nil {
		return err
	}
	return nil

}

func (s *PullRequestService) GetReviewPRs(userID string) ([]*entity.PullRequest, error) {
	if userID == "" {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "user_id cannot be empty",
		}
	}
	if len(userID) > config.MaxStringLength {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "user_id cannot exceed 255 characters",
		}
	}

	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "user not found",
		}
	}

	prs, err := s.prRepo.GetPRsByReviewer(userID)
	if err != nil {
		logging.Printf("ERROR: Failed to get PRs for reviewer %s: %v", userID, err)
		return nil, err
	}

	return prs, nil
}

func (s *PullRequestService) selectReviewers(candidates []*entity.User, n int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	if n > len(candidates) {
		n = len(candidates)
	}

	indices := make([]int, len(candidates))
	for i := range indices {
		indices[i] = i
	}

	for i := len(indices) - 1; i > 0; i-- {
		jBig, err := crand.Int(crand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			logging.Printf("ERROR: failed to generate crypto random: %v", err)
			break
		}
		j := int(jBig.Int64())
		indices[i], indices[j] = indices[j], indices[i]
	}

	result := make([]string, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, candidates[indices[i]].ID)
	}

	return result
}

func (s *PullRequestService) containsReviewer(reviewers []string, userID string) bool {
	for _, reviewer := range reviewers {
		if reviewer == userID {
			return true
		}
	}
	return false
}

func (s *PullRequestService) validateField(fieldName, value string) *entity.DomainError {
	if value == "" {
		return &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: fieldName + " cannot be empty",
		}
	}
	if len(value) > config.MaxStringLength {
		return &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: fieldName + " cannot exceed 255 characters",
		}
	}
	return nil
}

func (s *PullRequestService) buildReplacementCandidates(pr *entity.PullRequest, oldUserID string, reviewers []string) ([]*entity.User, error) {
	oldUser, err := s.userRepo.GetUser(oldUserID)
	if err != nil {
		logging.Printf("ERROR: Failed to get user %s: %v", oldUserID, err)
		return nil, err
	}
	if oldUser == nil {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "user not found",
		}
	}

	activeMembers, err := s.userRepo.GetActiveUsersByTeam(oldUser.Team)
	if err != nil {
		logging.Printf("ERROR: Failed to get active users for team %s: %v", oldUser.Team, err)
		return nil, err
	}

	candidates := make([]*entity.User, 0)
	for _, member := range activeMembers {
		if member.ID == oldUserID || member.ID == pr.AuthorID || s.containsReviewer(reviewers, member.ID) {
			continue
		}
		candidates = append(candidates, member)
	}
	return candidates, nil
}

func (s *PullRequestService) getAuthorAndCandidates(authorID string) ([]*entity.User, error) {
	author, err := s.userRepo.GetUser(authorID)
	if err != nil {
		logging.Printf("ERROR: Failed to get author %s: %v", authorID, err)
		return nil, err
	}
	if author == nil {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "author not found",
		}
	}

	team, err := s.teamRepo.GetTeam(author.Team)
	if err != nil {
		logging.Printf("ERROR: Failed to get team %s: %v", author.Team, err)
		return nil, err
	}
	if team == nil {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "team not found",
		}
	}

	activeMembers, err := s.userRepo.GetActiveUsersByTeam(author.Team)
	if err != nil {
		logging.Printf("ERROR: Failed to get active users for team %s: %v", author.Team, err)
		return nil, err
	}

	candidates := make([]*entity.User, 0)
	for _, member := range activeMembers {
		if member.ID != authorID {
			candidates = append(candidates, member)
		}
	}

	return candidates, nil
}
