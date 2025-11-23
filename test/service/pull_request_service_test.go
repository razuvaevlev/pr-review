
package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"pr-review/internal/entity"
	"pr-review/internal/service"
)

var longID = strings.Repeat("a", 256)


func makeMembers(ids ...string) []*entity.User {
	res := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		res = append(res, &entity.User{ID: id, Team: "team1", IsActive: true})
	}
	return res
}

func TestPullRequestService_CreatePR(t *testing.T) {

	tests := []struct {
		name     string
		prID     string
		prName   string
		authorID string
		prRepo   *mockPRRepo
		userRepo *mockUserRepo
		teamRepo *mockTeamRepo
		wantErr  bool
		errMsg   string
	}{
		{name: "empty_prid", prID: "", wantErr: true, errMsg: "pull_request_id cannot be empty"},
		{name: "too_long_prid", prID: longID, wantErr: true, errMsg: "cannot exceed 255"},
		{name: "empty_prname", prID: "p1", prName: "", wantErr: true, errMsg: "pull_request_name cannot be empty"},
		{name: "empty_author", prID: "p1", prName: "n1", authorID: "", wantErr: true, errMsg: "author_id cannot be empty"},
		{name: "pr_exists_error", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, errors.New("exists check err") }}, wantErr: true, errMsg: "exists check err"},
		{name: "pr_already_exists", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return true, nil }}, wantErr: true, errMsg: "PR id already exists"},
		{name: "author_get_error", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }}, userRepo: &mockUserRepo{GetUserFn: func(string) (*entity.User, error) { return nil, errors.New("user get err") }}, wantErr: true, errMsg: "user get err"},
		{name: "author_not_found", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }}, userRepo: &mockUserRepo{GetUserFn: func(string) (*entity.User, error) { return nil, nil }}, wantErr: true, errMsg: "author not found"},
		{name: "team_get_error", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }}, userRepo: &mockUserRepo{GetUserFn: func(string) (*entity.User, error) { return &entity.User{ID: "a1", Team: "team1"}, nil }}, teamRepo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return nil, errors.New("team err") }}, wantErr: true, errMsg: "team err"},
		{name: "team_not_found", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }}, userRepo: &mockUserRepo{GetUserFn: func(string) (*entity.User, error) { return &entity.User{ID: "a1", Team: "team1"}, nil }}, teamRepo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return nil, nil }}, wantErr: true, errMsg: "team not found"},
		{name: "active_members_error", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }}, userRepo: &mockUserRepo{GetUserFn: func(string) (*entity.User, error) { return &entity.User{ID: "a1", Team: "team1"}, nil }, GetActiveUsersByTeamFn: func(string) ([]*entity.User, error) { return nil, errors.New("active err") }}, teamRepo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return &entity.Team{Name: "team1"}, nil }}, wantErr: true, errMsg: "active err"},
		{name: "create_pr_error", prID: "p1", prName: "n1", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }, CreatePRFn: func(*entity.PullRequest) error { return errors.New("create pr failed") }}, userRepo: &mockUserRepo{GetUserFn: func(string) (*entity.User, error) { return &entity.User{ID: "a1", Team: "team1"}, nil }, GetActiveUsersByTeamFn: func(string) ([]*entity.User, error) { return makeMembers("a1", "r1"), nil }}, teamRepo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return &entity.Team{Name: "team1"}, nil }}, wantErr: true, errMsg: "create pr failed"},
		{name: "success", prID: "p2", prName: "n2", authorID: "a1", prRepo: &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }, CreatePRFn: func(*entity.PullRequest) error { return nil }}, userRepo: &mockUserRepo{GetUserFn: func(string) (*entity.User, error) { return &entity.User{ID: "a1", Team: "team1"}, nil }, GetActiveUsersByTeamFn: func(string) ([]*entity.User, error) { return makeMembers("a1", "r1", "r2"), nil }}, teamRepo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return &entity.Team{Name: "team1"}, nil }}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := tt.prRepo
			if prRepo == nil {
				prRepo = &mockPRRepo{PRExistsFn: func(string) (bool, error) { return false, nil }}
			}
			ur := tt.userRepo
			if ur == nil {
				ur = &mockUserRepo{}
			}
			tr := tt.teamRepo
			if tr == nil {
				tr = &mockTeamRepo{}
			}

			svc := service.NewPullRequestService(prRepo, ur, tr)

			pr, err := svc.CreatePR(tt.prID, tt.prName, tt.authorID)

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
			if pr == nil {
				t.Fatalf("expected pr, got nil")
			}
			if pr.ID != tt.prID {
				t.Fatalf("expected pr id %s, got %s", tt.prID, pr.ID)
			}
		})
	}
}

func TestPullRequestService_MergePR(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		prID       string
		prRepo     *mockPRRepo
		wantErr    bool
		errMsg     string
		wantMerged bool
	}{
		{name: "empty_prid", prID: "", wantErr: true, errMsg: "pull_request_id cannot be empty"},
		{name: "too_long_prid", prID: strings.Repeat("a", 256), wantErr: true, errMsg: "cannot exceed 255"},
		{name: "get_error", prID: "p1", prRepo: &mockPRRepo{GetPRFn: func(string) (*entity.PullRequest, error) { return nil, errors.New("get err") }}, wantErr: true, errMsg: "get err"},
		{name: "not_found", prID: "p2", prRepo: &mockPRRepo{GetPRFn: func(string) (*entity.PullRequest, error) { return nil, nil }}, wantErr: true, errMsg: "PR not found"},
		{name: "already_merged", prID: "p3", prRepo: &mockPRRepo{GetPRFn: func(string) (*entity.PullRequest, error) {
			return &entity.PullRequest{ID: "p3", Status: entity.StatusMerged, MergedAt: &now}, nil
		}}, wantErr: false, wantMerged: true},
		{name: "update_error", prID: "p4", prRepo: &mockPRRepo{GetPRFn: func(string) (*entity.PullRequest, error) {
			return &entity.PullRequest{ID: "p4", Status: entity.StatusOpen}, nil
		}, UpdatePRFn: func(*entity.PullRequest) error { return errors.New("update err") }}, wantErr: true, errMsg: "update err"},
		{name: "success", prID: "p5", prRepo: &mockPRRepo{GetPRFn: func(string) (*entity.PullRequest, error) {
			return &entity.PullRequest{ID: "p5", Status: entity.StatusOpen}, nil
		}, UpdatePRFn: func(*entity.PullRequest) error { return nil }}, wantErr: false, wantMerged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := tt.prRepo
			if prRepo == nil {
				prRepo = &mockPRRepo{}
			}
			svc := service.NewPullRequestService(prRepo, &mockUserRepo{}, &mockTeamRepo{})

			pr, err := svc.MergePR(tt.prID)

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
			if tt.wantMerged {
				if pr == nil {
					t.Fatalf("expected pr, got nil")
				}
				if pr.Status != entity.StatusMerged {
					t.Fatalf("expected merged status, got %v", pr.Status)
				}
				if pr.MergedAt == nil {
					t.Fatalf("expected MergedAt to be set")
				}
			}
		})
	}
}
