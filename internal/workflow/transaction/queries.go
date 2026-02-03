package transaction

import "go.temporal.io/sdk/workflow"

const (
	QueryTransactionState = "transaction.state"
)

func RegisterQueries(ctx workflow.Context, state *TransactionState) {
	err := workflow.SetQueryHandler(ctx, QueryTransactionState, func() (TransactionState, error) {
		return *state, nil
	})
	if err != nil {
		panic(err)
	}
}
