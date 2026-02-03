package models

import "time"

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeTopup    TransactionType = "TOPUP"
	TransactionTypePurchase TransactionType = "PURCHASE"
	TransactionTypeRefund   TransactionType = "REFUND"

	TransactionStatusPending  TransactionStatus = "PENDING"
	TransactionStatusSuccess  TransactionStatus = "SUCCESS"
	TransactionStatusFailed   TransactionStatus = "FAILED"
	TransactionStatusReversed TransactionStatus = "REVERSED"
)

type Transaction struct {
	ID          int64
	UserID      int64
	Amount      int64 // 👈 cents
	Type        TransactionType
	Status      TransactionStatus
	Reference   string
	Description string

	AdditionalInfo *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Transaction) TableName() string {
	return "transaction"
}

// Request DTO
type CreateTransactionRequest struct {
	Referance   string `json:"referance" validate:"required"`
	UserID      int64  `json:"user_id" validate:"required"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Type        string `json:"transaction_type" validate:"required,oneof=TOPUP PURCHASE"`
	Description string `json:"description" validate:"required"`

	Token string `json:"-"`
}

// Status Update DTO
type UpdateTransactionStatus struct {
	Reference string `validate:"required"`
	Status    string `validate:"required,oneof=PENDING SUCCESS FAILED REVERSED"`
	Reason    string
}
