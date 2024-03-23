package services

import (
	"context"
	"errors"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/database"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserIsAlreadyRegistered = errors.New("user is already registered")
	ErrUserIsNotExist          = errors.New("user is not exist")
	ErrPasswordIsIncorrect     = errors.New("password is incorrect")
)

type AuthService struct {
	storage Storage
}

type Storage interface {
	SaveUser(ctx context.Context, user models.UserWithHash) error

	FindUser(ctx context.Context, login string) (*models.UserWithHash, error)
}

func NewAuthService(storage Storage) *AuthService {
	return &AuthService{storage}
}

func (auth *AuthService) Register(ctx context.Context, user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	if err := auth.storage.SaveUser(ctx, models.UserWithHash{Login: *user.Login, Hash: string(hashedPassword)}); err != nil {
		if errors.Is(err, database.ErrDuplicateUser) {
			return ErrUserIsAlreadyRegistered
		}

		return err
	}

	return nil
}

func (auth *AuthService) Login(ctx context.Context, user models.User) error {
	u, err := auth.storage.FindUser(ctx, *user.Login)

	if err != nil {
		return err
	}

	if u == nil {
		return ErrUserIsNotExist
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(*user.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordIsIncorrect
		}

		return err
	}

	return nil
}

func (auth *AuthService) IsLoginValid(ctx context.Context, login string) error {
	user, err := auth.storage.FindUser(ctx, login)

	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserIsNotExist
	}

	return nil
}
