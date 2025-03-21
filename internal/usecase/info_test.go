package usecase

import (
	"avito-merch/internal/entity"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetTransfersByUsername(ctx context.Context, username string) ([]entity.Transaction, error) {
	args := m.Called(ctx, username)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func TestInfoUseCase_GetUserInfo_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTransactionRepo := new(MockTransactionRepository)

	mockUserRepo.On("GetUserByUsername", mock.Anything, "testuser").
		Return(&entity.User{
			Name:  "testuser",
			Coins: 1000,
		}, nil)

	mockUserRepo.On("GetUserInventory", mock.Anything, "testuser").
		Return([]entity.InventoryItem{
			{Type: "t-shirt", Quantity: 1},
		}, nil)

	mockTransactionRepo.On("GetTransfersByUsername", mock.Anything, "testuser").
		Return([]entity.Transaction{
			{FromUser: "user1", ToUser: "testuser", Amount: 100},
			{FromUser: "testuser", ToUser: "user2", Amount: 50},
		}, nil)

	uc := NewInfoUseCase(mockUserRepo, mockTransactionRepo)

	ctx := context.Background()
	username := "testuser"

	info, err := uc.GetUserInfo(ctx, username)

	assert.NoError(t, err)
	assert.Equal(t, 1000, info.Coins)
	assert.Equal(t, 1, len(info.Inventory))
	assert.Equal(t, 1, len(info.CoinHistory.Received))
	assert.Equal(t, 1, len(info.CoinHistory.Sent))

	mockUserRepo.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}
