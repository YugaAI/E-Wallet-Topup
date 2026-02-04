package repository

import (
	"context"
	"ewallet-topup/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepo struct {
	DB *gorm.DB
}

func (r *TransactionRepo) Create(ctx context.Context, trx *models.Transaction) error {
	return r.DB.WithContext(ctx).Create(trx).Error
}

func (r *TransactionRepo) FindByReference(ctx context.Context, ref string) (*models.Transaction, error) {

	var trx models.Transaction
	err := r.DB.WithContext(ctx).Where("reference = ?", ref).First(&trx).Error

	return &trx, err
}

func (r *TransactionRepo) FindByReferenceForUpdate(ctx context.Context, ref string) (*models.Transaction, error) {

	var trx models.Transaction
	err := r.DB.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("reference = ?", ref).First(&trx).Error

	return &trx, err
}

func (r *TransactionRepo) UpdateStatus(ctx context.Context, reference string, status models.TransactionStatus, reason *string) error {

	updateData := map[string]interface{}{
		"status": status,
	}

	if reason != nil {
		updateData["additional_info"] = *reason
	}

	return r.DB.WithContext(ctx).
		Model(&models.Transaction{}).
		Where("reference = ?", reference).
		Updates(updateData).
		Error
}
