package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"ewallet-topup/helpers"
	"ewallet-topup/internal/interfaces"
	"ewallet-topup/internal/models"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/activity"
)

type TransactionActivities struct {
	Service  interfaces.ITransactionService
	External interfaces.IExternal
}

type WalletRequest struct {
	Reference string  `json:"reference"`
	Amount    float64 `json:"amount"`
	UserID    int64   `json:"user_id,omitempty"`
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

func (a *TransactionActivities) DebitWallet(ctx context.Context, trx models.Transaction) error {

	logger := activity.GetLogger(ctx)
	logger.Info("credit wallet started", "reference", trx.Reference)

	walletHost := os.Getenv("WALLET_HOST")
	endpoint := os.Getenv("WALLET_ENDPOINT_DEBIT")
	url := strings.TrimRight(walletHost, "/") + "/" + strings.TrimLeft(endpoint, "/")

	reqBody := WalletRequest{
		Reference: trx.Reference,
		Amount:    trx.Amount,
		UserID:    trx.UserID,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal wallet request")
		return err
	}

	log.Debug().
		Str("url", url).
		Str("reference", trx.Reference).
		Str("token", maskToken(trx.Token)).
		Msg("sending request to wallet service")

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		log.Error().Err(err).Msg("failed to create http request")
		return err
	}

	req.Header.Set("Authorization", "Bearer "+trx.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("wallet service request failed")
		return fmt.Errorf("wallet service request failed: %w", err)
	}
	defer resp.Body.Close()

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		log.Error().Err(err).Msg("failed to decode wallet service response")
		return fmt.Errorf("failed to decode wallet service response: %w", err)
	}

	log.Debug().
		Int("status", resp.StatusCode).
		Interface("body", respBody).
		Msg("wallet service response")

	if resp.StatusCode == 401 {
		return fmt.Errorf("wallet service unauthorized: %d", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("wallet service error: %d - %v", resp.StatusCode, respBody)
	}

	log.Info().Str("reference", trx.Reference).Msg("DebitWallet activity success")
	return nil
}

func (a *TransactionActivities) CreditWallet(ctx context.Context, trx models.Transaction, token string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("credit wallet started", "reference", trx.Reference)

	// Ambil host & endpoint dari env
	walletHost := helpers.GetEnv("WALLET_HOST", "")
	if walletHost == "" {
		walletHost = "http://localhost:8085" // fallback default
	}
	endpoint := helpers.GetEnv("WALLET_ENDPOINT_CREDIT", "")
	if endpoint == "" {
		endpoint = "/wallet/v1/balance/credit"
	}
	url := strings.TrimRight(walletHost, "/") + "/" + strings.TrimLeft(endpoint, "/")

	// request body
	body := WalletRequest{
		Reference: trx.Reference,
		Amount:    trx.Amount,
		UserID:    trx.UserID,
	}
	data, err := json.Marshal(body)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal wallet request")
		return err
	}
	log.Debug().
		Str("url", url).
		Str("reference", trx.Reference).
		Str("token", maskToken(token)).
		Msg("sending request to wallet service")

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		log.Error().Err(err).Msg("failed to create http request")
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("wallet service request failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		log.Error().Err(err).Msg("failed to decode wallet service response")
		return fmt.Errorf("failed to decode wallet service response: %w", err)
	}

	log.Debug().
		Int("status", resp.StatusCode).
		Interface("body", respBody).
		Msg("wallet service response")

	if resp.StatusCode == 401 {
		logger.Error("unauthorized: token invalid or expired")
		return fmt.Errorf("wallet service unauthorized: 401")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("wallet service error: %d - %s", resp.StatusCode, respBody)
	}

	logger.Info("credit wallet success", "reference", trx.Reference)
	return nil
}

func (a *TransactionActivities) SendNotification(ctx context.Context, trx *models.Transaction, user models.TokenData) error {

	logger := activity.GetLogger(ctx)
	logger.Info("send success notification", "reference", trx.Reference)

	// never fail workflow because of notification
	a.Service.SendNotification(ctx, trx, user)

	return nil
}
