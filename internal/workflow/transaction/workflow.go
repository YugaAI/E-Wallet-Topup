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

	state := TransactionState{
		Reference: req.Referance,
		Status:    models.TransactionStatusPending,
		Step:      "INIT",
	}

	RegisterQueries(ctx, &state)
	ctx = workflow.WithActivityOptions(ctx, workflows.DefaultActivityOptions())

	logger.Info("transaction workflow started", "reference", req.Referance)

	var trx models.Transaction
	// STEP 1: create pending transaction (persist)
	if err := workflow.ExecuteActivity(
		ctx,
		(*TransactionActivities).CreatePendingTransaction,
		req,
	).Get(ctx, &trx); err != nil {
		state.Step = "CREATE_PANDING_FAILED"
		state.Status = models.TransactionStatusFailed
		return err
	}
	state.Step = "PENDING_CREATE"

	// STEP 2: wait confirmation (signal)
	confirmed := false
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(
		workflow.GetSignalChannel(ctx, SignalTransactionConfirm),
		func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			confirmed = true
		},
	)

	selector.AddReceive(
		workflow.GetSignalChannel(ctx, SignalTransactionCancel),
		func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			confirmed = false
		},
	)

	selector.Select(ctx)

	if !confirmed {
		state.Step = "CANCELLED"
		state.Status = models.TransactionStatusFailed

		_ = workflow.ExecuteActivity(
			ctx,
			(*TransactionActivities).UpdateTransactionStatus,
			trx.Reference,
			models.TransactionStatusFailed,
			nil,
		)
		return nil
	}

	state.Step = "CONFIRMED"

	// STEP 3: wallet operation
	if trx.Type == models.TransactionTypePurchase {
		if err := workflow.ExecuteActivity(
			ctx,
			(*TransactionActivities).DebitWallet,
			trx,
			req.Token,
		).Get(ctx, nil); err != nil {
			state.Step = "DEBIT_FAILED"
			state.Status = models.TransactionStatusFailed
			return err
		}
	}

	if trx.Type == models.TransactionTypeTopup {
		if err := workflow.ExecuteActivity(
			ctx,
			(*TransactionActivities).CreditWallet,
			trx,
			req.Token,
		).Get(ctx, nil); err != nil {
			state.Step = "CREDIT_FAILED"
			state.Status = models.TransactionStatusFailed
			return err
		}
	}

	// STEP 4: update status success
	if err := workflow.ExecuteActivity(
		ctx,
		(*TransactionActivities).UpdateTransactionStatus,
		trx.Reference,
		models.TransactionStatusSuccess,
		nil,
	).Get(ctx, nil); err != nil {
		state.Step = "UPDATE_STATUS_FAILED"
		return err
	}

	state.Step = "SUCCESS"
	state.Status = models.TransactionStatusSuccess

	// STEP 5: send notification (non blocking)
	if err := workflow.ExecuteActivity(
		ctx,
		(*TransactionActivities).SendSuccessNotification,
		trx,
		req.UserID,
	).Get(ctx, nil); err != nil {
		state.Step = "SEND_SUCCESS"
		return err
	}
	return nil
}
