package models

var transactionStatusFlow = map[TransactionStatus][]TransactionStatus{
	TransactionStatusPending: {
		TransactionStatusSuccess,
		TransactionStatusFailed,
		TransactionStatusReversed,
	},
	TransactionStatusSuccess: {
		TransactionStatusReversed,
	},
	TransactionStatusFailed:   {},
	TransactionStatusReversed: {},
}

func IsValidTransition(from TransactionStatus, to TransactionStatus) bool {
	nextStatuses, ok := transactionStatusFlow[from]
	if !ok {
		return false
	}
	for _, nextStatus := range nextStatuses {
		if to == nextStatus {
			return true
		}
	}
	return false
}

func (s TransactionStatus) IsFinal() bool {
	return s == TransactionStatusSuccess ||
		s == TransactionStatusFailed ||
		s == TransactionStatusReversed
}
