package transactiontypes

type TransactionType int
type TransactionPurpose int

const (
	Credit      = TransactionType(1)
	Debit       = TransactionType(2)
	Award       = TransactionPurpose(1)
	Redeem      = TransactionPurpose(2)
	TransferOut = TransactionPurpose(3)
	TransferIn  = TransactionPurpose(4)
)
