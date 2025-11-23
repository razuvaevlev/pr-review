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

var _ repo.TeamRepository = (*TeamRepository)(nil)

type TeamRepository struct {
	db  *pgxpool.Pool
	sb  squirrel.StatementBuilderType
	ctx context.Context
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{
		db:  db,
		sb:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		ctx: context.Background(),
	}
}

func (r *TeamRepository) CreateTeam(team *entity.Team) error {
	if team == nil {
		return errors.New("team cannot be nil")
	}
	if team.Name == "" {
		return errors.New("team_name cannot be empty")
	}
	if len(team.Name) > config.MaxStringLength {
		return errors.New("team_name cannot exceed 255 characters")
	}

	query := r.sb.Insert("teams").
		Columns("team_name").
		Values(team.Name)

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for CreateTeam: %v", err)
		return err
	}

	_, err = r.db.Exec(r.ctx, sql, args...)
	if err != nil {
		logging.Printf("ERROR: Failed to execute CreateTeam query for team %s: %v", team.Name, err)
		return err
	}

	return nil
}

func (r *TeamRepository) GetTeam(teamName string) (*entity.Team, error) {
	if teamName == "" {
		return nil, errors.New("team_name cannot be empty")
	}
	if len(teamName) > config.MaxStringLength {
		return nil, errors.New("team_name cannot exceed 255 characters")
	}

	exists, err := r.TeamExists(teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	query := r.sb.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(squirrel.Eq{"team_name": teamName})

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for GetTeam: %v", err)
		return nil, err
	}

	rows, err := r.db.Query(r.ctx, sql, args...)
	if err != nil {
		logging.Printf("ERROR: Failed to execute GetTeam query for team %s: %v", teamName, err)
		return nil, err
	}
	defer rows.Close()

	members := make([]entity.User, 0)
	for rows.Next() {
		var user entity.User
		err := rows.Scan(&user.ID, &user.Name, &user.Team, &user.IsActive)
		if err != nil {
			return nil, err
		}
		members = append(members, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &entity.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

func (r *TeamRepository) TeamExists(teamName string) (bool, error) {
	if teamName == "" {
		return false, errors.New("team_name cannot be empty")
	}
	if len(teamName) > config.MaxStringLength {
		return false, errors.New("team_name cannot exceed 255 characters")
	}

	query := r.sb.Select("COUNT(*)").
		From("teams").
		Where(squirrel.Eq{"team_name": teamName})

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for TeamExists: %v", err)
		return false, err
	}

	var count int
	err = r.db.QueryRow(r.ctx, sql, args...).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		logging.Printf("ERROR: Failed to execute TeamExists query for team %s: %v", teamName, err)
		return false, err
	}

	return count > 0, nil
}
