package interfaces

import (
	"context"

	"github.com/Ramsi97/flowra-back-end/internal/auth/domain"
)

// AuthRepository defines data-access operations for authentication.
type AuthRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
}
