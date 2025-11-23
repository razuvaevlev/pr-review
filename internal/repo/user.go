package repo

import "pr-review/internal/entity"

type UserRepository interface {
	GetUser(userID string) (*entity.User, error)

	CreateOrUpdateUser(user *entity.User) error

	UpdateUser(user *entity.User) error

	GetUsersByTeam(teamName string) ([]*entity.User, error)

	GetActiveUsersByTeam(teamName string) ([]*entity.User, error)
}
