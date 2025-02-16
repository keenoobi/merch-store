package usecase

import (
	"avito-merch/internal/entity"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserInventory(ctx context.Context, username string) ([]entity.InventoryItem, error) {
	args := m.Called(ctx, username)
	return args.Get(0).([]entity.InventoryItem), args.Error(1)
}

func (m *MockUserRepository) UpdateUserAfterPurchase(ctx context.Context, username string, amount int) error {
	args := m.Called(ctx, username, amount)
	return args.Error(0)
}

func TestAuthUseCase_Authenticate_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	mockUserRepo.On("GetUserByUsername", mock.Anything, "testuser").
		Return((*entity.User)(nil), nil)

	mockUserRepo.On("Create", mock.Anything, &entity.User{
		Name:     "testuser",
		Password: "password",
		Coins:    1000,
	}).
		Return(nil)

	uc := NewAuthUseCase(mockUserRepo)

	ctx := context.Background()
	username := "testuser"
	password := "password"

	user, err := uc.Authenticate(ctx, username, password)

	assert.NoError(t, err)
	assert.Equal(t, username, user.Name)
	assert.Equal(t, password, user.Password)
	assert.Equal(t, 1000, user.Coins)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthUseCase_Authenticate_CreateUserError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	mockUserRepo.On("GetUserByUsername", mock.Anything, "testuser").
		Return((*entity.User)(nil), nil)

	mockUserRepo.On("Create", mock.Anything, &entity.User{
		Name:     "testuser",
		Password: "password",
		Coins:    1000,
	}).
		Return(errors.New("database error"))

	uc := NewAuthUseCase(mockUserRepo)

	ctx := context.Background()
	username := "testuser"
	password := "password"

	user, err := uc.Authenticate(ctx, username, password)

	assert.Error(t, err)
	assert.Nil(t, user)

	mockUserRepo.AssertExpectations(t)
}
