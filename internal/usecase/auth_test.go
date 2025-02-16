package usecase

import (
	"avito-merch/internal/entity"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository - мок для UserRepository
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
	// Создаем мок UserRepository
	mockUserRepo := new(MockUserRepository)

	// Настраиваем ожидания
	mockUserRepo.On("GetUserByUsername", mock.Anything, "testuser").
		Return((*entity.User)(nil), nil) // Пользователь не найден

	mockUserRepo.On("Create", mock.Anything, &entity.User{
		Name:     "testuser",
		Password: "password",
		Coins:    10000000,
	}).
		Return(nil) // Успешное создание пользователя

	// Создаем usecase с моком
	uc := NewAuthUseCase(mockUserRepo)

	ctx := context.Background()
	username := "testuser"
	password := "password"

	// Вызываем метод Authenticate
	user, err := uc.Authenticate(ctx, username, password)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, username, user.Name)
	assert.Equal(t, password, user.Password)
	assert.Equal(t, 10000000, user.Coins)

	// Проверяем, что методы были вызваны
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUseCase_Authenticate_CreateUserError(t *testing.T) {
	// Создаем мок UserRepository
	mockUserRepo := new(MockUserRepository)

	// Настраиваем ожидания
	mockUserRepo.On("GetUserByUsername", mock.Anything, "testuser").
		Return((*entity.User)(nil), nil) // Пользователь не найден

	mockUserRepo.On("Create", mock.Anything, &entity.User{
		Name:     "testuser",
		Password: "password",
		Coins:    10000000,
	}).
		Return(errors.New("database error")) // Ошибка при создании пользователя

	// Создаем usecase с моком
	uc := NewAuthUseCase(mockUserRepo)

	ctx := context.Background()
	username := "testuser"
	password := "password"

	// Вызываем метод Authenticate
	user, err := uc.Authenticate(ctx, username, password)

	// Проверяем результат
	assert.Error(t, err)
	assert.Nil(t, user)

	// Проверяем, что методы были вызваны
	mockUserRepo.AssertExpectations(t)
}
