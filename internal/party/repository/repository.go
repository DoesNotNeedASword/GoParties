package repository

import (
	"context"
	"errors"
	"fmt"

	"Parties/internal/party/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrNotFound = errors.New("party not found")

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Repository struct {
	db DB
}

func New(db DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, party model.Party) (model.Party, error) {
	const query = `
		INSERT INTO parties (name, description, promises, image, is_allowed, ban_reason)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, promises, image, is_allowed, ban_reason`

	created, err := scanParty(r.db.QueryRow(
		ctx,
		query,
		party.Name,
		party.Description,
		party.Promises,
		party.Image,
		party.IsAllowed,
		textFromStringPtr(party.BanReason),
	))
	if err != nil {
		return model.Party{}, fmt.Errorf("create party: %w", err)
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (model.Party, error) {
	const query = `
		SELECT id, name, description, promises, image, is_allowed, ban_reason
		FROM parties
		WHERE id = $1`

	party, err := scanParty(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return model.Party{}, fmt.Errorf("get party by id: %w", err)
	}

	return party, nil
}

func (r *Repository) List(ctx context.Context, filter model.ListPartiesFilter) ([]model.Party, error) {
	const query = `
		SELECT id, name, description, promises, image, is_allowed, ban_reason
		FROM parties
		ORDER BY id
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("list parties: %w", err)
	}
	defer rows.Close()

	parties := make([]model.Party, 0)
	for rows.Next() {
		var party model.Party
		var banReason pgtype.Text

		if err := rows.Scan(
			&party.ID,
			&party.Name,
			&party.Description,
			&party.Promises,
			&party.Image,
			&party.IsAllowed,
			&banReason,
		); err != nil {
			return nil, fmt.Errorf("scan party: %w", err)
		}
		if banReason.Valid {
			party.BanReason = &banReason.String
		}
		parties = append(parties, party)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate parties: %w", err)
	}

	return parties, nil
}

func (r *Repository) Update(ctx context.Context, party model.Party) (model.Party, error) {
	const query = `
		UPDATE parties
		SET name = $2,
			description = $3,
			promises = $4,
			image = $5,
			is_allowed = $6,
			ban_reason = $7
		WHERE id = $1
		RETURNING id, name, description, promises, image, is_allowed, ban_reason`

	updated, err := scanParty(r.db.QueryRow(
		ctx,
		query,
		party.ID,
		party.Name,
		party.Description,
		party.Promises,
		party.Image,
		party.IsAllowed,
		textFromStringPtr(party.BanReason),
	))
	if err != nil {
		return model.Party{}, fmt.Errorf("update party: %w", err)
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM parties WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete party: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func scanParty(row pgx.Row) (model.Party, error) {
	var party model.Party
	var banReason pgtype.Text

	if err := row.Scan(
		&party.ID,
		&party.Name,
		&party.Description,
		&party.Promises,
		&party.Image,
		&party.IsAllowed,
		&banReason,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Party{}, ErrNotFound
		}
		return model.Party{}, err
	}

	if banReason.Valid {
		party.BanReason = &banReason.String
	}

	return party, nil
}

func textFromStringPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}
