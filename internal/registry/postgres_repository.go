package registry

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgUniqueViolation = "23505"
	pgInvalidTextRepr = "22P02" // malformed uuid
	serviceColumns    = `id::text, name, team, description, tags, health_url, check_interval_seconds, timeout_seconds, expected_status, created_at, updated_at`
	listWhere         = `deleted_at IS NULL AND ($1 = '' OR team = $1) AND ($2 = '' OR tags @> ARRAY[$2]::text[])`
)

// PostgresRepository stores services in the services table.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a Repository backed by pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, in Input) (Service, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO services (name, team, description, tags, health_url, check_interval_seconds, timeout_seconds, expected_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING `+serviceColumns,
		in.Name, in.Team, in.Description, in.Tags, in.HealthURL, in.CheckIntervalSeconds, in.TimeoutSeconds, in.ExpectedStatus)
	return scanOne(row)
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Service, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+serviceColumns+` FROM services WHERE id = $1::uuid AND deleted_at IS NULL`, id)
	return scanOne(row)
}

func (r *PostgresRepository) List(ctx context.Context, f ListFilter) ([]Service, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM services WHERE `+listWhere, f.Team, f.Tag).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count services: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+serviceColumns+` FROM services WHERE `+listWhere+` ORDER BY created_at, id LIMIT $3 OFFSET $4`,
		f.Team, f.Tag, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	items := make([]Service, 0, f.Limit)
	for rows.Next() {
		s, err := scanRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan service: %w", err)
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list services: %w", err)
	}
	return items, total, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id string, in Input) (Service, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE services
		SET name = $2, team = $3, description = $4, tags = $5, health_url = $6,
		    check_interval_seconds = $7, timeout_seconds = $8, expected_status = $9, updated_at = now()
		WHERE id = $1::uuid AND deleted_at IS NULL
		RETURNING `+serviceColumns,
		id, in.Name, in.Team, in.Description, in.Tags, in.HealthURL, in.CheckIntervalSeconds, in.TimeoutSeconds, in.ExpectedStatus)
	return scanOne(row)
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE services SET deleted_at = now(), updated_at = now() WHERE id = $1::uuid AND deleted_at IS NULL`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanOne(row pgx.Row) (Service, error) {
	s, err := scanRow(row)
	if err != nil {
		return Service{}, mapError(err)
	}
	return s, nil
}

func scanRow(row pgx.Row) (Service, error) {
	var s Service
	err := row.Scan(&s.ID, &s.Name, &s.Team, &s.Description, &s.Tags, &s.HealthURL,
		&s.CheckIntervalSeconds, &s.TimeoutSeconds, &s.ExpectedStatus, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

// mapError turns driver errors into the repository's domain errors.
func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return ErrNameTaken
		case pgInvalidTextRepr:
			return ErrNotFound
		}
	}
	return err
}
