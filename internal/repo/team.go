package repo

import "pr-review/internal/entity"

type TeamRepository interface {
	CreateTeam(team *entity.Team) error

	GetTeam(teamName string) (*entity.Team, error)

	TeamExists(teamName string) (bool, error)
}
