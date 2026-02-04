package transaction

type SignalTransaction struct {
	Reason *string
}

const (
	SignalTransactionConfirm = "transacation.confirm"
	SignalTransactionCancel  = "transacation.cancel"
)
