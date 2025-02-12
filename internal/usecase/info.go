// internal/usecase/info.go
package usecase

import (
	"avito-merch/internal/entity"
	"avito-merch/internal/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type InfoUseCase struct {
	userRepo        *repository.UserRepository
	transactionRepo *repository.TransactionRepository
}

func NewInfoUseCase(userRepo *repository.UserRepository, transactionRepo *repository.TransactionRepository) *InfoUseCase {
	return &InfoUseCase{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (uc *InfoUseCase) GetUserInfo(ctx context.Context, userID uuid.UUID) (*entity.InfoData, error) {
	info, err := uc.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	return info, nil
}
