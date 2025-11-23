package service

import (
	"pr-review/internal/config"
	"pr-review/internal/entity"
	"pr-review/internal/logging"
	"pr-review/internal/repo"
)

type UserService struct {
	userRepo  repo.UserRepository
	prService *PullRequestService
}

func NewUserService(userRepo repo.UserRepository, prService *PullRequestService) *UserService {
	return &UserService{
		userRepo:  userRepo,
		prService: prService,
	}
}

func (s *UserService) SetIsActive(userID string, isActive bool) (*entity.User, error) {
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

	user.IsActive = isActive
	if err := s.userRepo.UpdateUser(user); err != nil {
		logging.Printf("ERROR: Failed to update user %s: %v", userID, err)
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetReviewPRs(userID string) ([]*entity.PullRequest, error) {
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
		logging.Printf("ERROR: Failed to get user %s: %v", userID, err)
		return nil, err
	}
	if user == nil {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "user not found",
		}
	}

	return s.prService.GetReviewPRs(userID)
}
