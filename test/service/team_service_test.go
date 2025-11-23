package service_test

import (
	"errors"
	"strings"
	"testing"

	"pr-review/internal/entity"
	"pr-review/internal/service"
)

var longName = strings.Repeat("a", 256)

var validMember = entity.User{ID: "u1", Name: "n1", Team: "team1", IsActive: true}


type mockTeamRepo struct {
	CreateTeamFn func(*entity.Team) error
	GetTeamFn    func(string) (*entity.Team, error)
	TeamExistsFn func(string) (bool, error)
}

func (m *mockTeamRepo) CreateTeam(team *entity.Team) error {
	if m.CreateTeamFn != nil {
		return m.CreateTeamFn(team)
	}
	return nil
}
func (m *mockTeamRepo) GetTeam(name string) (*entity.Team, error) {
	if m.GetTeamFn != nil {
		return m.GetTeamFn(name)
	}
	return nil, nil
}
func (m *mockTeamRepo) TeamExists(name string) (bool, error) {
	if m.TeamExistsFn != nil {
		return m.TeamExistsFn(name)
	}
	return false, nil
}

func TestTeamService_AddTeam(t *testing.T) {

	tests := []struct {
		name     string
		team     *entity.Team
		teamRepo *mockTeamRepo
		userRepo *mockUserRepo
		wantErr  bool
		errMsg   string
	}{
		{name: "empty_name", team: &entity.Team{Name: "", Members: []entity.User{validMember}}, wantErr: true, errMsg: "team_name cannot be empty"},
		{name: "too_long_name", team: &entity.Team{Name: longName, Members: []entity.User{validMember}}, wantErr: true, errMsg: "cannot exceed 255"},
		{name: "no_members", team: &entity.Team{Name: "team1", Members: []entity.User{}}, wantErr: true, errMsg: "team must have at least one member"},
		{name: "member_empty_id", team: &entity.Team{Name: "team1", Members: []entity.User{{ID: "", Name: "n", Team: "team1"}}}, wantErr: true, errMsg: "member user_id cannot be empty"},
		{name: "member_id_too_long", team: &entity.Team{Name: "team1", Members: []entity.User{{ID: longName, Name: "n", Team: "team1"}}}, wantErr: true, errMsg: "member user_id cannot exceed 255"},
		{name: "member_name_empty", team: &entity.Team{Name: "team1", Members: []entity.User{{ID: "u", Name: "", Team: "team1"}}}, wantErr: true, errMsg: "member username cannot be empty"},
		{name: "member_name_too_long", team: &entity.Team{Name: "team1", Members: []entity.User{{ID: "u", Name: longName, Team: "team1"}}}, wantErr: true, errMsg: "member username cannot exceed 255"},
		{name: "member_team_mismatch", team: &entity.Team{Name: "team1", Members: []entity.User{{ID: "u", Name: "n", Team: "other"}}}, wantErr: true, errMsg: "member team_name must match team name"},
		{name: "team_exists_check_error", team: &entity.Team{Name: "team1", Members: []entity.User{validMember}}, teamRepo: &mockTeamRepo{TeamExistsFn: func(string) (bool, error) { return false, errors.New("exists err") }}, wantErr: true, errMsg: "exists err"},
		{name: "team_already_exists", team: &entity.Team{Name: "team1", Members: []entity.User{validMember}}, teamRepo: &mockTeamRepo{TeamExistsFn: func(string) (bool, error) { return true, nil }}, wantErr: true, errMsg: "team_name already exists"},
		{name: "user_create_error", team: &entity.Team{Name: "team1", Members: []entity.User{validMember}}, teamRepo: &mockTeamRepo{TeamExistsFn: func(string) (bool, error) { return false, nil }}, userRepo: &mockUserRepo{CreateOrUpdateUserFn: func(_ *entity.User) error { return errors.New("create user failed") }}, wantErr: true, errMsg: "create user failed"},
		{name: "create_team_error", team: &entity.Team{Name: "team1", Members: []entity.User{validMember}}, teamRepo: &mockTeamRepo{TeamExistsFn: func(string) (bool, error) { return false, nil }, CreateTeamFn: func(_ *entity.Team) error { return errors.New("create team failed") }}, wantErr: true, errMsg: "create team failed"},
		{name: "success", team: &entity.Team{Name: "team1", Members: []entity.User{validMember}}, teamRepo: &mockTeamRepo{TeamExistsFn: func(string) (bool, error) { return false, nil }}, userRepo: &mockUserRepo{CreateOrUpdateUserFn: func(_ *entity.User) error { return nil }}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := tt.teamRepo
			if tr == nil {
				tr = &mockTeamRepo{TeamExistsFn: func(string) (bool, error) { return false, nil }}
			}
			ur := tt.userRepo
			if ur == nil {
				ur = &mockUserRepo{}
			}

			svc := service.NewTeamService(tr, ur)

			err := svc.AddTeam(tt.team)
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
		})
	}
}

func TestTeamService_GetTeam(t *testing.T) {
	longName := strings.Repeat("a", 256)

	tests := []struct {
		name     string
		teamName string
		repo     *mockTeamRepo
		wantErr  bool
		errMsg   string
	}{
		{name: "empty_name", teamName: "", wantErr: true, errMsg: "team_name cannot be empty"},
		{name: "too_long_name", teamName: longName, wantErr: true, errMsg: "cannot exceed 255"},
		{name: "repo_error", teamName: "t1", repo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return nil, errors.New("get err") }}, wantErr: true, errMsg: "get err"},
		{name: "not_found", teamName: "t2", repo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return nil, nil }}, wantErr: true, errMsg: "team not found"},
		{name: "success", teamName: "t3", repo: &mockTeamRepo{GetTeamFn: func(string) (*entity.Team, error) { return &entity.Team{Name: "t3"}, nil }}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &mockTeamRepo{}
			}
			svc := service.NewTeamService(repo, &mockUserRepo{})

			tm, err := svc.GetTeam(tt.teamName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if derr, ok := err.(*entity.DomainError); ok {
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
			if tm == nil {
				t.Fatalf("expected team, got nil")
			}
		})
	}
}
