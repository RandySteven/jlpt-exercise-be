package enums

type TransactionType string

const (
	TransactionTypeSubscription  TransactionType = "subscription"
	TransactionTypeClassPurchase TransactionType = "class_purchase"
)

const (
	TransactionTypePayment TransactionType = "payment"
	TransactionTypeRefund  TransactionType = "refund"
)
