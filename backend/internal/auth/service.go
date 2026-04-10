package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo         *Repository
	tokenManager *TokenManager
}

func NewService(repo *Repository, tokenManager *TokenManager) *Service {
	return &Service{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return AuthResponse{}, apperrors.NewValidation(fields)
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return AuthResponse{}, apperrors.WrapInternal(err)
	}

	user := User{
		ID:       uuid.NewString(),
		Name:     strings.TrimSpace(req.Name),
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Password: hash,
	}

	created, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return AuthResponse{}, apperrors.NewValidation(map[string]string{
				"email": "already in use",
			})
		}
		return AuthResponse{}, apperrors.WrapInternal(err)
	}

	token, err := s.tokenManager.Issue(created)
	if err != nil {
		return AuthResponse{}, apperrors.WrapInternal(err)
	}

	return AuthResponse{
		Token: token,
		User:  scrubPassword(created),
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return AuthResponse{}, apperrors.NewValidation(fields)
	}

	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AuthResponse{}, apperrors.NewUnauthorized()
		}
		return AuthResponse{}, apperrors.WrapInternal(err)
	}

	if err := CheckPassword(user.Password, req.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return AuthResponse{}, apperrors.NewUnauthorized()
		}
		return AuthResponse{}, apperrors.WrapInternal(fmt.Errorf("compare password: %w", err))
	}

	token, err := s.tokenManager.Issue(user)
	if err != nil {
		return AuthResponse{}, apperrors.WrapInternal(err)
	}

	return AuthResponse{
		Token: token,
		User:  scrubPassword(user),
	}, nil
}

func scrubPassword(user User) User {
	user.Password = ""
	return user
}
