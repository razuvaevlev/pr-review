package service

import (
	"pr-review/internal/config"
	"pr-review/internal/entity"
	"pr-review/internal/logging"
	"pr-review/internal/repo"
)

type TeamService struct {
	teamRepo repo.TeamRepository
	userRepo repo.UserRepository
}

func NewTeamService(teamRepo repo.TeamRepository, userRepo repo.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) AddTeam(team *entity.Team) error {
	if team.Name == "" {
		return &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "team_name cannot be empty",
		}
	}
	if len(team.Name) > config.MaxStringLength {
		return &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "team_name cannot exceed 255 characters",
		}
	}

	if len(team.Members) == 0 {
		return &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "team must have at least one member",
		}
	}

	for _, member := range team.Members {
		if derr := s.validateTeamMember(&member, team.Name); derr != nil {
			return derr
		}
	}

	exists, err := s.teamRepo.TeamExists(team.Name)
	if err != nil {
		logging.Printf("ERROR: Failed to check if team exists: %v", err)
		return err
	}
	if exists {
		return &entity.DomainError{
			Code:    entity.ErrorCodeTeamExists,
			Message: "team_name already exists",
		}
	}

	if err := s.teamRepo.CreateTeam(team); err != nil {
		logging.Printf("ERROR: Failed to create team %s: %v", team.Name, err)
		return err
	}

	for _, member := range team.Members {
		user := &entity.User{
			ID:       member.ID,
			Name:     member.Name,
			Team:     team.Name,
			IsActive: member.IsActive,
		}
		if err := s.userRepo.CreateOrUpdateUser(user); err != nil {
			logging.Printf("ERROR: Failed to create/update user %s: %v", user.ID, err)
			return err
		}
	}

	return nil
}

func (s *TeamService) validateTeamMember(member *entity.User, teamName string) *entity.DomainError {
	if member.ID == "" {
		return &entity.DomainError{Code: entity.ErrorCodeNotFound, Message: "member user_id cannot be empty"}
	}
	if len(member.ID) > config.MaxStringLength {
		return &entity.DomainError{Code: entity.ErrorCodeNotFound, Message: "member user_id cannot exceed 255 characters"}
	}
	if member.Name == "" {
		return &entity.DomainError{Code: entity.ErrorCodeNotFound, Message: "member username cannot be empty"}
	}
	if len(member.Name) > config.MaxStringLength {
		return &entity.DomainError{Code: entity.ErrorCodeNotFound, Message: "member username cannot exceed 255 characters"}
	}
	if member.Team != teamName {
		return &entity.DomainError{Code: entity.ErrorCodeNotFound, Message: "member team_name must match team name"}
	}
	return nil
}

func (s *TeamService) GetTeam(teamName string) (*entity.Team, error) {
	if teamName == "" {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "team_name cannot be empty",
		}
	}
	if len(teamName) > config.MaxStringLength {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "team_name cannot exceed 255 characters",
		}
	}

	team, err := s.teamRepo.GetTeam(teamName)
	if err != nil {
		logging.Printf("ERROR: Failed to get team %s: %v", teamName, err)
		return nil, err
	}
	if team == nil {
		return nil, &entity.DomainError{
			Code:    entity.ErrorCodeNotFound,
			Message: "team not found",
		}
	}
	return team, nil
}
