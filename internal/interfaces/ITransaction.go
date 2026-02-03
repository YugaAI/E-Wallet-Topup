package interfaces

import (
	"context"
	"ewallet-topup/internal/models"

	"github.com/gin-gonic/gin"
)

type ITransactionRepo interface {
	Create(ctx context.Context, trx *models.Transaction) error
	FindByReference(ctx context.Context, ref string) (*models.Transaction, error)
	FindByReferenceForUpdate(ctx context.Context, ref string) (*models.Transaction, error)
	UpdateStatus(ctx context.Context, ref string, status models.TransactionStatus, reason *string) error
}

type ITransactionService interface {
	UpdateStatus(ctx context.Context, ref string, status models.TransactionStatus, reason *string) error
	CreatePending(ctx context.Context, req models.CreateTransactionRequest) (*models.Transaction, error)
	DebitWallet(ctx context.Context, trx *models.Transaction, token string) error
	CreditWallet(ctx context.Context, trx *models.Transaction, token string) error
	SendNotification(ctx context.Context, trx *models.Transaction, user models.TokenData)
}

type ITransactionAPI interface {
	CreateTransaction(c *gin.Context)
	UpdateStatusTransaction(c *gin.Context)
	//RefundTransaction(c *gin.Context)
	//GetTransaction(c *gin.Context)
	//GetTransactionDetail(c *gin.Context)
}
