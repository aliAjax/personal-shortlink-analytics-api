package service

import (
	"context"
	"errors"
	"testing"

	"github.com/example/shortlink-api/internal/user/model"
)

type fakeRepository struct {
	createErr error
	findErr   error
}

func (f *fakeRepository) Create(context.Context, string, string) (model.User, error) {
	return model.User{}, f.createErr
}
func (f *fakeRepository) FindByUsername(context.Context, string) (model.User, error) {
	return model.User{}, f.findErr
}
func (f *fakeRepository) FindByID(context.Context, int64) (model.User, error) {
	return model.User{}, f.findErr
}

func TestRegisterPreservesRepositoryError(t *testing.T) {
	sentinel := errors.New("duplicate account")
	_, err := NewService(&fakeRepository{createErr: sentinel}).Register(context.Background(), model.RegisterRequest{
		Username: "alice",
		Password: "password",
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("repository error lost identity: %v", err)
	}
}

func TestLoginPreservesUnexpectedRepositoryError(t *testing.T) {
	sentinel := errors.New("connection reset")
	_, err := NewService(&fakeRepository{findErr: sentinel}).Login(context.Background(), model.LoginRequest{Username: "alice", Password: "password"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("repository error lost identity: %v", err)
	}
}
