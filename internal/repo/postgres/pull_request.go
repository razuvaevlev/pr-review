package postgres

import (
	"context"
	"errors"
	"fmt"
	"pr-review/internal/config"
	"pr-review/internal/entity"
	"pr-review/internal/logging"
	"pr-review/internal/repo"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repo.PullRequestRepository = (*PullRequestRepository)(nil)

type PullRequestRepository struct {
	db  *pgxpool.Pool
	sb  squirrel.StatementBuilderType
	ctx context.Context
}

func NewPullRequestRepository(db *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{
		db:  db,
		sb:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		ctx: context.Background(),
	}
}

func (r *PullRequestRepository) validatePR(pr *entity.PullRequest) error {
	if pr == nil {
		return errors.New("pull request cannot be nil")
	}
	if pr.ID == "" {
		return errors.New("pull_request_id cannot be empty")
	}
	if len(pr.ID) > config.MaxStringLength {
		return errors.New("pull_request_id cannot exceed 255 characters")
	}
	if pr.Name == "" {
		return errors.New("pull_request_name cannot be empty")
	}
	if len(pr.Name) > config.MaxStringLength {
		return errors.New("pull_request_name cannot exceed 255 characters")
	}
	if pr.AuthorID == "" {
		return errors.New("author_id cannot be empty")
	}
	if len(pr.AuthorID) > config.MaxStringLength {
		return errors.New("author_id cannot exceed 255 characters")
	}
	if pr.Status != entity.StatusOpen && pr.Status != entity.StatusMerged {
		return fmt.Errorf("invalid status: %s", pr.Status)
	}
	return nil
}

func (r *PullRequestRepository) validatePRID(prID string) error {
	if prID == "" {
		return errors.New("pull_request_id cannot be empty")
	}
	if len(prID) > config.MaxStringLength {
		return errors.New("pull_request_id cannot exceed 255 characters")
	}
	return nil
}

func (r *PullRequestRepository) executeInTransaction(operation func(tx pgx.Tx) error, operationName string) error {
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(r.ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logging.Printf("ERROR: failed to rollback transaction in %s: %v", operationName, err)
		}
	}()

	if err := operation(tx); err != nil {
		return err
	}

	return tx.Commit(r.ctx)
}

func (r *PullRequestRepository) insertPRData(tx pgx.Tx, pr *entity.PullRequest) error {
	query := r.sb.Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at").
		Values(
			pr.ID,
			pr.Name,
			pr.AuthorID,
			string(pr.Status),
			pr.CreatedAt,
			pr.MergedAt,
		)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(r.ctx, sql, args...)
	return err
}

func (r *PullRequestRepository) updatePRData(tx pgx.Tx, pr *entity.PullRequest) error {
	query := r.sb.Update("pull_requests").
		Set("pull_request_name", pr.Name).
		Set("author_id", pr.AuthorID).
		Set("status", string(pr.Status)).
		Set("created_at", pr.CreatedAt).
		Set("merged_at", pr.MergedAt).
		Where(squirrel.Eq{"pull_request_id": pr.ID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(r.ctx, sql, args...)
	return err
}

func (r *PullRequestRepository) CreatePR(pr *entity.PullRequest) error {
	if err := r.validatePR(pr); err != nil {
		return err
	}

	return r.executeInTransaction(func(tx pgx.Tx) error {
		if err := r.insertPRData(tx, pr); err != nil {
			return err
		}

		if len(pr.AssignedReviewers) > 0 {
			return r.insertReviewers(tx, pr.ID, pr.AssignedReviewers)
		}
		return nil
	}, "CreatePR")
}

func (r *PullRequestRepository) GetPR(prID string) (*entity.PullRequest, error) {
	if err := r.validatePRID(prID); err != nil {
		return nil, err
	}

	query := r.sb.Select(
		"pull_request_id",
		"pull_request_name",
		"author_id",
		"status",
		"created_at",
		"merged_at",
	).
		From("pull_requests").
		Where(squirrel.Eq{"pull_request_id": prID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var pr entity.PullRequest
	var statusStr string
	var createdAt, mergedAt *time.Time

	err = r.db.QueryRow(r.ctx, sql, args...).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&statusStr,
		&createdAt,
		&mergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	pr.Status = entity.Status(statusStr)
	pr.CreatedAt = createdAt
	pr.MergedAt = mergedAt

	reviewers, err := r.getReviewers(prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PullRequestRepository) UpdatePR(pr *entity.PullRequest) error {
	if err := r.validatePR(pr); err != nil {
		return err
	}

	return r.executeInTransaction(func(tx pgx.Tx) error {
		if err := r.updatePRData(tx, pr); err != nil {
			return err
		}

		if _, err := tx.Exec(r.ctx, "DELETE FROM assigned_reviewers WHERE pull_request_id = $1", pr.ID); err != nil {
			return err
		}

		if len(pr.AssignedReviewers) > 0 {
			return r.insertReviewers(tx, pr.ID, pr.AssignedReviewers)
		}
		return nil
	}, "UpdatePR")
}

func (r *PullRequestRepository) PRExists(prID string) (bool, error) {
	if err := r.validatePRID(prID); err != nil {
		return false, err
	}

	query := r.sb.Select("COUNT(*)").
		From("pull_requests").
		Where(squirrel.Eq{"pull_request_id": prID})

	sql, args, err := query.ToSql()
	if err != nil {
		logging.Printf("ERROR: Failed to build SQL query for PRExists: %v", err)
		return false, err
	}

	var count int
	err = r.db.QueryRow(r.ctx, sql, args...).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		logging.Printf("ERROR: Failed to execute PRExists query for PR %s: %v", prID, err)
		return false, err
	}

	return count > 0, nil
}

func (r *PullRequestRepository) scanPR(scanner interface{ Scan(...interface{}) error }) (*entity.PullRequest, error) {
	var pr entity.PullRequest
	var statusStr string
	var createdAt, mergedAt *time.Time

	if err := scanner.Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&statusStr,
		&createdAt,
		&mergedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	pr.Status = entity.Status(statusStr)
	pr.CreatedAt = createdAt
	pr.MergedAt = mergedAt

	reviewers, err := r.getReviewers(pr.ID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PullRequestRepository) validateUserID(userID string) error {
	if userID == "" {
		return errors.New("user_id cannot be empty")
	}
	if len(userID) > config.MaxStringLength {
		return errors.New("user_id cannot exceed 255 characters")
	}
	return nil
}

func (r *PullRequestRepository) GetPRsByReviewer(userID string) ([]*entity.PullRequest, error) {
	if err := r.validateUserID(userID); err != nil {
		return nil, err
	}

	query := r.sb.Select(
		"pr.pull_request_id",
		"pr.pull_request_name",
		"pr.author_id",
		"pr.status",
		"pr.created_at",
		"pr.merged_at",
	).
		From("pull_requests pr").
		Join("assigned_reviewers ar ON pr.pull_request_id = ar.pull_request_id").
		Where(squirrel.Eq{"ar.reviewer_id": userID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(r.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prs := make([]*entity.PullRequest, 0)
	for rows.Next() {
		pr, err := r.scanPR(rows)
		if err != nil {
			logging.Printf("ERROR: Failed to scan PR row: %v", err)
			return nil, err
		}
		if pr == nil {
			continue
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

func (r *PullRequestRepository) insertReviewers(tx pgx.Tx, prID string, reviewers []string) error {
	if len(reviewers) == 0 {
		return nil
	}

	query := r.sb.Insert("assigned_reviewers").
		Columns("pull_request_id", "reviewer_id")

	for _, reviewer := range reviewers {
		if reviewer != "" {
			query = query.Values(prID, reviewer)
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(r.ctx, sql, args...)
	return err
}

func (r *PullRequestRepository) getReviewers(prID string) ([]string, error) {
	query := r.sb.Select("reviewer_id").
		From("assigned_reviewers").
		Where(squirrel.Eq{"pull_request_id": prID}).
		OrderBy("assigned_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(r.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var reviewer string
		if err := rows.Scan(&reviewer); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

func (r *PullRequestRepository) AddReviewers(prID string, reviewers []string) error {
	if prID == "" {
		return errors.New("pull_request_id cannot be empty")
	}
	if len(reviewers) == 0 {
		return nil
	}

	exists, err := r.PRExists(prID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("pull request with id %s does not exist", prID)
	}

	return r.insertReviewers(nil, prID, reviewers)
}

func (r *PullRequestRepository) RemoveReviewers(prID string, reviewers []string) error {
	if prID == "" {
		return errors.New("pull_request_id cannot be empty")
	}
	if len(reviewers) == 0 {
		return nil
	}

	query := r.sb.Delete("assigned_reviewers").
		Where(squirrel.Eq{"pull_request_id": prID, "reviewer_id": reviewers})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(r.ctx, sql, args...)
	return err
}
