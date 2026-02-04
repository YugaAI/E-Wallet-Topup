// workflow/transaction_workflow.go
package transaction

import (
	"ewallet-topup/internal/models"
	workflows "ewallet-topup/internal/workflow"

	"go.temporal.io/sdk/workflow"
)

type TransactionState struct {
	Reference string
	Status    models.TransactionStatus
	Step      string
	Reason    *string
}

func TransactionWorkflow(ctx workflow.Context, req models.CreateTransactionRequest) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("transaction workflow started", "reference", req.Referance)

	state := TransactionState{
		Reference: req.Referance,
		Status:    models.TransactionStatusPending,
		Step:      "INIT",
	}

	RegisterQueries(ctx, &state)
	ctx = workflow.WithActivityOptions(ctx, workflows.DefaultActivityOptions())

	var trx models.Transaction
	// STEP 1: create pending transaction
	logger.Debug("executing CreatePendingTransaction activity", "reference", req.Referance)
	if err := workflow.ExecuteActivity(ctx, (*TransactionActivities).CreatePendingTransaction, req).Get(ctx, &trx); err != nil {
		state.Step = "CREATE_PENDING_FAILED"
		state.Status = models.TransactionStatusFailed
		logger.Error("CreatePendingTransaction failed", "error", err)
		return err
	}
	state.Step = "PENDING_CREATE"
	logger.Info("pending transaction created", "reference", trx.Reference)

	// STEP 2: wait for confirmation
	logger.Debug("waiting for confirmation signal", "reference", trx.Reference)
	confirmed := false
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(workflow.GetSignalChannel(ctx, SignalTransactionConfirm), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, nil)
		confirmed = true
		logger.Info("transaction confirmed signal received", "reference", trx.Reference)
	})

	selector.AddReceive(workflow.GetSignalChannel(ctx, SignalTransactionCancel), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, nil)
		confirmed = false
		logger.Info("transaction cancel signal received", "reference", trx.Reference)
	})

	selector.Select(ctx)

	if !confirmed {
		state.Step = "CANCELLED"
		state.Status = models.TransactionStatusFailed

		_ = workflow.ExecuteActivity(ctx, (*TransactionActivities).UpdateTransactionStatus, trx.Reference, models.TransactionStatusFailed, nil).Get(ctx, nil)
		return nil
	}

	state.Step = "CONFIRMED"

	// STEP 3: wallet operation
	if trx.Type == models.TransactionTypeTopup {
		logger.Debug("executing CreditWallet activity", "reference", trx.Reference, "token", maskToken(req.Token))
		var walletResp interface{}
		err := workflow.ExecuteActivity(ctx, (*TransactionActivities).CreditWallet, trx, req.Token).Get(ctx, &walletResp)
		if err != nil {
			state.Step = "CREDIT_FAILED"
			state.Status = models.TransactionStatusFailed
			logger.Error("CreditWallet failed", "error", err)
			return err
		}
		logger.Info("CreditWallet activity success", "reference", trx.Reference)
	}

	// STEP 4: update status success
	if err := workflow.ExecuteActivity(ctx, (*TransactionActivities).UpdateTransactionStatus, trx.Reference, models.TransactionStatusSuccess, nil).Get(ctx, nil); err != nil {
		state.Step = "UPDATE_STATUS_FAILED"
		logger.Error("UpdateTransactionStatus failed", "error", err)
		return err
	}
	state.Step = "SUCCESS"
	state.Status = models.TransactionStatusSuccess

	// STEP 5: send notification
	if err := workflow.ExecuteActivity(ctx, (*TransactionActivities).SendNotification, trx, models.TokenData{UserID: req.UserID}).Get(ctx, nil); err != nil {
		state.Step = "SEND_NOTIFICATION_FAILED"
		logger.Error("SendNotification failed", "error", err)
		return err
	}

	return nil
}

// mask token supaya aman di log
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
