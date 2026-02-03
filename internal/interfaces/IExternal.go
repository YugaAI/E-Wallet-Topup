package interfaces

import (
	"context"
	"ewallet-topup/external"
	"ewallet-topup/internal/models"
)

type IExternal interface {
	ValidateToken(ctx context.Context, token string) (models.TokenData, error)
	CreditBalance(ctx context.Context, token string, req external.UpdateBalance) (*external.UpdateBalanceResponse, error)
	DebitBalance(ctx context.Context, token string, req external.UpdateBalance) (*external.UpdateBalanceResponse, error)
	NotifyUserRegistered(userID int64, email, fullName string) error
}
