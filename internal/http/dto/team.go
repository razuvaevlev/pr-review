package dto

import (
	"errors"
	"pr-review/internal/config"
	"pr-review/internal/entity"
	"strconv"
	"strings"
)

type TeamRequest struct {
	TeamName string          `json:"team_name" binding:"required"`
	Members  []MemberRequest `json:"members" binding:"required"`
}

func (t *TeamRequest) Validate() error {
	if strings.TrimSpace(t.TeamName) == "" {
		return errors.New("team_name cannot be empty")
	}
	if len(t.TeamName) > config.MaxStringLength {
		return errors.New("team_name cannot exceed 255 characters")
	}
	if len(t.Members) == 0 {
		return errors.New("members cannot be empty")
	}
	if len(t.Members) > config.MaxTeamMembers {
		return errors.New("members cannot exceed 100")
	}

	for i, member := range t.Members {
		if err := member.Validate(); err != nil {
			return errors.New("member[" + strconv.Itoa(i) + "]: " + err.Error())
		}
	}

	return nil
}

type MemberRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	IsActive bool   `json:"is_active" binding:"required"`
}

func (m *MemberRequest) Validate() error {
	if strings.TrimSpace(m.UserID) == "" {
		return errors.New("user_id cannot be empty")
	}
	if len(m.UserID) > config.MaxStringLength {
		return errors.New("user_id cannot exceed 255 characters")
	}
	if strings.TrimSpace(m.Username) == "" {
		return errors.New("username cannot be empty")
	}
	if len(m.Username) > config.MaxStringLength {
		return errors.New("username cannot exceed 255 characters")
	}
	return nil
}

func (t *TeamRequest) ToEntity() *entity.Team {
	members := make([]entity.User, len(t.Members))
	for i, m := range t.Members {
		members[i] = entity.User{
			ID:       m.UserID,
			Name:     m.Username,
			Team:     t.TeamName,
			IsActive: m.IsActive,
		}
	}

	return &entity.Team{
		Name:    t.TeamName,
		Members: members,
	}
}
