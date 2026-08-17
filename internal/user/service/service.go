package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/example/shortlink-api/internal/user/model"
	"github.com/example/shortlink-api/internal/user/repository"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, req model.RegisterRequest) (model.User, error) {
	username := strings.TrimSpace(req.Username)
	if len(username) < 3 || len(username) > 64 {
		return model.User{}, errors.New("username must be 3 to 64 characters")
	}
	if len(req.Password) < 6 {
		return model.User{}, errors.New("password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	user, err := s.repo.Create(ctx, username, string(hash))
	if err != nil {
		return model.User{}, fmt.Errorf("register user: %w", err)
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, req model.LoginRequest) (model.User, error) {
	user, err := s.repo.FindByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, ErrInvalidCredentials
		}
		return model.User{}, fmt.Errorf("find user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return model.User{}, ErrInvalidCredentials
	}
	return user, nil
}
