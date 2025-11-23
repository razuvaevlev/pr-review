package postgres

import (
	"context"
	"errors"
	"pr-review/internal/config"
	"pr-review/internal/entity"
	"pr-review/internal/logging"
	"pr-review/internal/repo"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repo.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db  *pgxpool.Pool
	sb  squirrel.StatementBuilderType
	ctx context.Context
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db:  db,
		sb:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		ctx: context.Background(),
	}
}

func (r *UserRepository) GetUser(userID string) (*entity.User, error) {
	if userID == "" {
		return nil, errors.New("user_id cannot be empty")
	}
	if len(userID) > config.MaxStringLength {
		return nil, errors.New("user_id cannot exceed 255 characters")
	}

	query := r.sb.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(squirrel.Eq{"user_id": userID})

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for GetUser: %v", err)
		return nil, err
	}

	var user entity.User
	err = r.db.QueryRow(r.ctx, sql, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Team,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		logging.Printf("ERROR: Failed to execute GetUser query for user %s: %v", userID, err)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreateOrUpdateUser(user *entity.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	if user.ID == "" {
		return errors.New("user_id cannot be empty")
	}
	if len(user.ID) > config.MaxStringLength {
		return errors.New("user_id cannot exceed 255 characters")
	}
	if user.Name == "" {
		return errors.New("username cannot be empty")
	}
	if len(user.Name) > config.MaxStringLength {
		return errors.New("username cannot exceed 255 characters")
	}
	if user.Team == "" {
		return errors.New("team_name cannot be empty")
	}
	if len(user.Team) > config.MaxStringLength {
		return errors.New("team_name cannot exceed 255 characters")
	}

	query := r.sb.Insert("users").
		Columns("user_id", "username", "team_name", "is_active").
		Values(user.ID, user.Name, user.Team, user.IsActive).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active")

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for CreateOrUpdateUser: %v", err)
		return err
	}

	_, err = r.db.Exec(r.ctx, sql, args...)
	if err != nil {
		logging.Printf("ERROR: Failed to execute CreateOrUpdateUser query for user %s: %v", user.ID, err)
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(user *entity.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	if user.ID == "" {
		return errors.New("user_id cannot be empty")
	}
	if len(user.ID) > config.MaxStringLength {
		return errors.New("user_id cannot exceed 255 characters")
	}
	if user.Name == "" {
		return errors.New("username cannot be empty")
	}
	if len(user.Name) > config.MaxStringLength {
		return errors.New("username cannot exceed 255 characters")
	}
	if user.Team == "" {
		return errors.New("team_name cannot be empty")
	}
	if len(user.Team) > config.MaxStringLength {
		return errors.New("team_name cannot exceed 255 characters")
	}

	query := r.sb.Update("users").
		Set("username", user.Name).
		Set("team_name", user.Team).
		Set("is_active", user.IsActive).
		Where(squirrel.Eq{"user_id": user.ID})

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for UpdateUser: %v", err)
		return err
	}

	_, err = r.db.Exec(r.ctx, sql, args...)
	if err != nil {
		logging.Printf("ERROR: Failed to execute UpdateUser query for user %s: %v", user.ID, err)
		return err
	}
	return nil
}

func (r *UserRepository) GetUsersByTeam(teamName string) ([]*entity.User, error) {
	if teamName == "" {
		return nil, errors.New("team_name cannot be empty")
	}
	if len(teamName) > config.MaxStringLength {
		return nil, errors.New("team_name cannot exceed 255 characters")
	}

	query := r.sb.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(squirrel.Eq{"team_name": teamName})

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for GetUsersByTeam: %v", err)
		return nil, err
	}

	rows, err := r.db.Query(r.ctx, sql, args...)
	if err != nil {
		logging.Printf("ERROR: Failed to execute GetUsersByTeam query for team %s: %v", teamName, err)
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		var user entity.User
		err := rows.Scan(&user.ID, &user.Name, &user.Team, &user.IsActive)
		if err != nil {
			logging.Printf("ERROR: Failed to scan user row: %v", err)
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		logging.Printf("ERROR: Error iterating rows in GetUsersByTeam: %v", err)
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetActiveUsersByTeam(teamName string) ([]*entity.User, error) {
	if teamName == "" {
		return nil, errors.New("team_name cannot be empty")
	}
	if len(teamName) > config.MaxStringLength {
		return nil, errors.New("team_name cannot exceed 255 characters")
	}

	query := r.sb.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(squirrel.Eq{"team_name": teamName, "is_active": true})

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for GetActiveUsersByTeam: %v", err)
		return nil, err
	}

	rows, err := r.db.Query(r.ctx, sql, args...)
	if err != nil {
		logging.Printf("ERROR: Failed to execute GetActiveUsersByTeam query for team %s: %v", teamName, err)
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		var user entity.User
		err := rows.Scan(&user.ID, &user.Name, &user.Team, &user.IsActive)
		if err != nil {
			logging.Printf("ERROR: Failed to scan user row: %v", err)
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		logging.Printf("ERROR: Error iterating rows in GetActiveUsersByTeam: %v", err)
		return nil, err
	}

	return users, nil
}
