
package service_test

import (
	"errors"
	"strings"
	"testing"

	"pr-review/internal/entity"
	"pr-review/internal/service"
)


type mockUserRepo struct {
	GetUserFn              func(string) (*entity.User, error)
	CreateOrUpdateUserFn   func(*entity.User) error
	UpdateUserFn           func(*entity.User) error
	GetUsersByTeamFn       func(string) ([]*entity.User, error)
	GetActiveUsersByTeamFn func(string) ([]*entity.User, error)
}

func (m *mockUserRepo) GetUser(userID string) (*entity.User, error) {
	if m.GetUserFn != nil {
		return m.GetUserFn(userID)
	}
	return nil, nil
}
func (m *mockUserRepo) CreateOrUpdateUser(user *entity.User) error {
	if m.CreateOrUpdateUserFn != nil {
		return m.CreateOrUpdateUserFn(user)
	}
	return nil
}
func (m *mockUserRepo) UpdateUser(user *entity.User) error {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(user)
	}
	return nil
}
func (m *mockUserRepo) GetUsersByTeam(teamName string) ([]*entity.User, error) {
	if m.GetUsersByTeamFn != nil {
		return m.GetUsersByTeamFn(teamName)
	}
	return nil, nil
}
func (m *mockUserRepo) GetActiveUsersByTeam(teamName string) ([]*entity.User, error) {
	if m.GetActiveUsersByTeamFn != nil {
		return m.GetActiveUsersByTeamFn(teamName)
	}
	return nil, nil
}


type mockPRRepo struct {
	CreatePRFn         func(*entity.PullRequest) error
	GetPRFn            func(string) (*entity.PullRequest, error)
	UpdatePRFn         func(*entity.PullRequest) error
	PRExistsFn         func(string) (bool, error)
	GetPRsByReviewerFn func(string) ([]*entity.PullRequest, error)
}

func (m *mockPRRepo) CreatePR(pr *entity.PullRequest) error {
	if m.CreatePRFn != nil {
		return m.CreatePRFn(pr)
	}
	return nil
}
func (m *mockPRRepo) GetPR(prID string) (*entity.PullRequest, error) {
	if m.GetPRFn != nil {
		return m.GetPRFn(prID)
	}
	return nil, nil
}
func (m *mockPRRepo) UpdatePR(pr *entity.PullRequest) error {
	if m.UpdatePRFn != nil {
		return m.UpdatePRFn(pr)
	}
	return nil
}
func (m *mockPRRepo) PRExists(prID string) (bool, error) {
	if m.PRExistsFn != nil {
		return m.PRExistsFn(prID)
	}
	return false, nil
}
func (m *mockPRRepo) GetPRsByReviewer(userID string) ([]*entity.PullRequest, error) {
	if m.GetPRsByReviewerFn != nil {
		return m.GetPRsByReviewerFn(userID)
	}
	return nil, nil
}

func TestUserService_SetIsActive(t *testing.T) {
	longID := strings.Repeat("a", 256)

	tests := []struct {
		name     string
		userID   string
		isActive bool
		repo     *mockUserRepo
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "empty_user_id",
			userID:  "",
			wantErr: true,
			errMsg:  "user_id cannot be empty",
		},
		{
			name:    "too_long_user_id",
			userID:  longID,
			wantErr: true,
			errMsg:  "cannot exceed 255",
		},
		{
			name:   "repo_get_error",
			userID: "u1",
			repo: &mockUserRepo{
				GetUserFn: func(_ string) (*entity.User, error) { return nil, errors.New("db error") },
			},
			wantErr: true,
			errMsg:  "db error",
		},
		{
			name:   "user_not_found",
			userID: "u2",
			repo: &mockUserRepo{
				GetUserFn: func(_ string) (*entity.User, error) { return nil, nil },
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:   "update_error",
			userID: "u3",
			repo: &mockUserRepo{
				GetUserFn:    func(_ string) (*entity.User, error) { return &entity.User{ID: "u3", IsActive: false}, nil },
				UpdateUserFn: func(_ *entity.User) error { return errors.New("update failed") },
			},
			wantErr: true,
			errMsg:  "update failed",
		},
		{
			name:     "success_update",
			userID:   "u4",
			isActive: true,
			repo: &mockUserRepo{
				GetUserFn: func(_ string) (*entity.User, error) { return &entity.User{ID: "u4", IsActive: false}, nil },
				UpdateUserFn: func(u *entity.User) error {
					if !u.IsActive {
						return errors.New("was not set active")
					}
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &mockUserRepo{}
			}

			svc := service.NewUserService(repo, nil)

			u, err := svc.SetIsActive(tt.userID, tt.isActive)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				
				var derr *entity.DomainError
				if errors.As(err, &derr) {
					if !strings.Contains(derr.Message, tt.errMsg) {
						t.Fatalf("unexpected domain error message: %v", derr.Message)
					}
				} else {
					if !strings.Contains(err.Error(), tt.errMsg) {
						t.Fatalf("unexpected error: %v", err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u == nil {
				t.Fatalf("expected user, got nil")
			}
			if u.IsActive != tt.isActive {
				t.Fatalf("expected IsActive=%v, got %v", tt.isActive, u.IsActive)
			}
		})
	}
}

func TestUserService_GetReviewPRs(t *testing.T) {
	longID := strings.Repeat("a", 256)

	tests := []struct {
		name     string
		userID   string
		userRepo *mockUserRepo
		prRepo   *mockPRRepo
		wantErr  bool
		errMsg   string
		wantLen  int
	}{
		{name: "empty_user_id", userID: "", wantErr: true, errMsg: "user_id cannot be empty"},
		{name: "too_long_user_id", userID: longID, wantErr: true, errMsg: "cannot exceed 255"},
		{name: "repo_get_error", userID: "u1", userRepo: &mockUserRepo{GetUserFn: func(_ string) (*entity.User, error) { return nil, errors.New("db get err") }}, wantErr: true, errMsg: "db get err"},
		{name: "user_not_found", userID: "u2", userRepo: &mockUserRepo{GetUserFn: func(_ string) (*entity.User, error) { return nil, nil }}, wantErr: true, errMsg: "user not found"},
		{name: "pr_service_error", userID: "u3", userRepo: &mockUserRepo{GetUserFn: func(_ string) (*entity.User, error) { return &entity.User{ID: "u3"}, nil }}, prRepo: &mockPRRepo{GetPRsByReviewerFn: func(_ string) ([]*entity.PullRequest, error) { return nil, errors.New("pr error") }}, wantErr: true, errMsg: "pr error"},
		{name: "success", userID: "u4", userRepo: &mockUserRepo{GetUserFn: func(_ string) (*entity.User, error) { return &entity.User{ID: "u4"}, nil }}, prRepo: &mockPRRepo{GetPRsByReviewerFn: func(_ string) ([]*entity.PullRequest, error) {
			return []*entity.PullRequest{{ID: "p1", Name: "PR1", AuthorID: "a1"}}, nil
		}}, wantErr: false, wantLen: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := tt.userRepo
			if userRepo == nil {
				userRepo = &mockUserRepo{}
			}
			prRepo := tt.prRepo
			if prRepo == nil {
				prRepo = &mockPRRepo{}
			}

			prService := service.NewPullRequestService(prRepo, userRepo, nil)
			svc := service.NewUserService(userRepo, prService)

			prs, err := svc.GetReviewPRs(tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				var derr *entity.DomainError
				if errors.As(err, &derr) {
					if !strings.Contains(derr.Message, tt.errMsg) {
						t.Fatalf("unexpected domain error message: %v", derr.Message)
					}
				} else {
					if !strings.Contains(err.Error(), tt.errMsg) {
						t.Fatalf("unexpected error: %v", err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(prs) != tt.wantLen {
				t.Fatalf("expected %d prs, got %d", tt.wantLen, len(prs))
			}
		})
	}
}
