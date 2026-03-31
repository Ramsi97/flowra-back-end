package usecase

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/auth/domain"
	"github.com/Ramsi97/flowra-back-end/internal/auth/repository/interfaces"
	"github.com/Ramsi97/flowra-back-end/pkg/hash"
	"github.com/Ramsi97/flowra-back-end/pkg/jwt"
)

type authusecase struct {
	repo      interfaces.AuthRepository
	jwtSecret string
}

// NewAuthUseCase constructs an AuthUseCase backed by the provided repository.
func NewAuthUseCase(repo interfaces.AuthRepository, jwtSecret string) domain.AuthUseCase {
	return &authusecase{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register validates input, hashes the password, and persists the user.
func (u *authusecase) Register(user *domain.User) error {
	if user.Email == "" || user.Password == "" || user.FullName == "" {
		return errors.New("full_name, email, and password are required")
	}

	// Email format validation
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(user.Email) {
		return errors.New("invalid email format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existing, err := u.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("email already registered")
	}

	hashed, err := hash.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashed

	return u.repo.CreateUser(ctx, user)
}

// Login verifies credentials and returns a signed JWT on success.
func (u *authusecase) Login(email, password string) (domain.UserResponse, error) {
	if email == "" || password == "" {
		return domain.UserResponse{}, errors.New("email and password are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.UserResponse{}, err
	}
	if user == nil || !hash.CheckPassword(user.Password, password) {
		return domain.UserResponse{}, errors.New("invalid email or password")
	}

	token, err := jwt.GenerateToken(user.ID, u.jwtSecret)
	if err != nil {
		return domain.UserResponse{}, err
	}

	// Clear the hash before returning the user in the response.
	user.Password = ""

	return domain.UserResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Logout is stateless — the client discards the token.
// This method can be extended with a token blocklist in the future.
func (u *authusecase) Logout() error {
	return nil
}
