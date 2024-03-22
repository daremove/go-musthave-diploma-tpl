package services

import (
	"context"
	"errors"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserIsAlreadyRegistered = errors.New("ErrUserIsAlreadyRegistered")
)

type AuthService struct {
	storage Storage
}

type Storage interface {
	SaveUser(ctx context.Context, login, hash string) error

	FindUser(ctx context.Context, login string) (*models.User, error)
}

func NewAuthService(storage Storage) *AuthService {
	return &AuthService{storage}
}

func (auth *AuthService) Register(ctx context.Context, user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	if err := auth.storage.SaveUser(ctx, *user.Login, string(hashedPassword)); err != nil {
		// todo check uniq constraint from db
		return ErrUserIsAlreadyRegistered
	}

	return nil
}

func (auth *AuthService) IsLoginValid(ctx context.Context, login string) (bool, error) {
	user, err := auth.storage.FindUser(ctx, login)

	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return true, nil
}
