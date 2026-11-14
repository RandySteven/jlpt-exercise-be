package enums

type PaymentType string

const (
	PaymentTypeSubscription PaymentType = "SUBSCRIPTION"
	PaymentTypeClass        PaymentType = "CLASS"
)

func (p PaymentType) ToString() string {
	return string(p)
}
