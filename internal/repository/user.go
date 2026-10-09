package repository

import (
	"context"
	"database/sql"
	"errors"

	"siakad-mini/internal/model"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository { return &UserRepository{} }

func (r *UserRepository) FindByEmail(ctx context.Context, q DBTX, email string) (*model.User, error) {
	return r.one(ctx, q, `SELECT id, email, password, role FROM users WHERE email = $1`, email)
}

func (r *UserRepository) FindByID(ctx context.Context, q DBTX, id int64) (*model.User, error) {
	return r.one(ctx, q, `SELECT id, email, password, role FROM users WHERE id = $1`, id)
}

func (r *UserRepository) one(ctx context.Context, q DBTX, query string, args ...any) (*model.User, error) {
	var u model.User
	err := q.QueryRowContext(ctx, query, args...).Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) EmailExists(ctx context.Context, q DBTX, email string) (bool, error) {
	var exists bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) Create(ctx context.Context, q DBTX, email, passwordHash, role string) (int64, error) {
	var id int64
	err := q.QueryRowContext(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`,
		email, passwordHash, role).Scan(&id)
	return id, err
}
