package transaction

import (
	"context"
	"ewallet-topup/internal/interfaces"
	"ewallet-topup/internal/models"
	"fmt"

	"go.temporal.io/sdk/activity"
)

type TransactionActivities struct {
	Service  interfaces.ITransactionService
	External interfaces.IExternal
}

func (a *TransactionActivities) CreatePendingTransaction(ctx context.Context, req models.CreateTransactionRequest) (*models.Transaction, error) {

	logger := activity.GetLogger(ctx)
	logger.Info("create pending transaction", "reference", req.Referance)

	trx, err := a.Service.CreatePending(ctx, req)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

func (a *TransactionActivities) UpdateTransactionStatus(ctx context.Context, ref string, status models.TransactionStatus, reason *string) error {

	logger := activity.GetLogger(ctx)
	logger.Info("update transaction status", "reference", ref, "status", status)

	return a.Service.UpdateStatus(ctx, ref, status, reason)
}

func (a *TransactionActivities) DebitWallet(ctx context.Context, trx *models.Transaction, token string) error {

	logger := activity.GetLogger(ctx)
	logger.Info("debit wallet", "reference", trx.Reference)

	if trx.Type != models.TransactionTypePurchase {
		return fmt.Errorf("invalid transaction type for debit: %s", trx.Type)
	}

	return a.Service.DebitWallet(ctx, trx, token)
}

func (a *TransactionActivities) CreditWallet(ctx context.Context, trx *models.Transaction, token string) error {

	logger := activity.GetLogger(ctx)
	logger.Info("credit wallet", "reference", trx.Reference)

	return a.Service.CreditWallet(ctx, trx, token)
}

func (a *TransactionActivities) SendSuccessNotification(ctx context.Context, trx *models.Transaction, user models.TokenData) error {

	logger := activity.GetLogger(ctx)
	logger.Info("send success notification", "reference", trx.Reference)

	// never fail workflow because of notification
	a.Service.SendNotification(ctx, trx, user)

	return nil
}
