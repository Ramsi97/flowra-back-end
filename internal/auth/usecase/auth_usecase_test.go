package usecase

import (
	"context"
	"testing"

	"github.com/Ramsi97/flowra-back-end/internal/auth/domain"
	"github.com/Ramsi97/flowra-back-end/internal/auth/repository/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Ensure MockAuthRepository implements AuthRepository
var _ interfaces.AuthRepository = (*MockAuthRepository)(nil)

func TestUpdateProfile(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	uc := NewAuthUseCase(mockRepo, "secret")

	userID := "user-123"
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		existingUser := &domain.User{
			ID:       userID,
			FullName: "Old Name",
			Email:    "test@example.com",
		}
		updateRequest := &domain.User{
			FullName: "New Name",
			Gender:   "Male",
		}

		mockRepo.On("FindByID", mock.Anything, userID).Return(existingUser, nil).Once()
		mockRepo.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.FullName == "New Name" && u.Gender == "Male" && u.ID == userID
		})).Return(nil).Once()

		err := uc.UpdateProfile(ctx, userID, updateRequest)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()

		err := uc.UpdateProfile(ctx, userID, &domain.User{FullName: "New Name"})

		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
