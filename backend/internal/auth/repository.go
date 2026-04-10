package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user User) (User, error) {
	const query = `
		INSERT INTO users (id, name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password, created_at
	`

	var created User
	err := r.db.QueryRow(ctx, query, user.ID, user.Name, user.Email, user.Password).Scan(
		&created.ID,
		&created.Name,
		&created.Email,
		&created.Password,
		&created.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	return created, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (User, error) {
	const query = `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}
