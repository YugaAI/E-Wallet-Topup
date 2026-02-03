package services

import (
	"context"
	"ewallet-topup/external"
	"ewallet-topup/internal/interfaces"
	"ewallet-topup/internal/models"
	"fmt"
)

type TransactionService struct {
	TransactionRepo interfaces.ITransactionRepo
	External        interfaces.IExternal
}

func NewTransactionService(Repo interfaces.ITransactionRepo, Ext interfaces.IExternal) *TransactionService {
	return &TransactionService{
		TransactionRepo: Repo,
		External:        Ext,
	}
}

func (s *TransactionService) CreatePending(ctx context.Context, req models.CreateTransactionRequest) (*models.Transaction, error) {

	trx := &models.Transaction{
		UserID:      req.UserID,
		Amount:      req.Amount,
		Type:        models.TransactionType(req.Type),
		Status:      models.TransactionStatusPending,
		Reference:   req.Referance,
		Description: req.Description,
	}

	err := s.TransactionRepo.Create(ctx, trx)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

func (s *TransactionService) UpdateStatus(ctx context.Context, ref string, status models.TransactionStatus, reason *string) error {
	trx, err := s.TransactionRepo.FindByReferenceForUpdate(ctx, ref)
	if err != nil {
		return err
	}
	if !models.IsValidTransition(trx.Status, status) {
		return fmt.Errorf("invalid status transition %s -> %s", trx.Status, status)
	}
	return s.TransactionRepo.UpdateStatus(ctx, ref, status, reason)
}

func (s *TransactionService) DebitWallet(ctx context.Context, trx *models.Transaction, token string) error {
	req := external.UpdateBalance{
		Amount:    float64(trx.Amount),
		Reference: trx.Reference,
	}

	_, err := s.External.DebitBalance(ctx, token, req)
	return err
}

func (s *TransactionService) CreditWallet(ctx context.Context, trx *models.Transaction, token string) error {
	req := external.UpdateBalance{
		Amount:    float64(trx.Amount),
		Reference: trx.Reference,
	}
	_, err := s.External.CreditBalance(ctx, token, req)
	return err
}

func (s *TransactionService) SendNotification(ctx context.Context, trx *models.Transaction, user models.TokenData) {
	if trx.Type != models.TransactionTypePurchase || trx.Status != models.TransactionStatusSuccess {
		return
	}

	err := s.External.NotifyUserRegistered(
		user.UserID,
		user.Email,
		user.FullName,
	)
	if err != nil {
		return
	}
}
