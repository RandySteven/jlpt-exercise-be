package enums

type (
	MidtransStatus string
)

func (m MidtransStatus) String() string {
	return string(m)
}

const (
	MidtransPendingStatus           MidtransStatus = `pending`
	MidtransCaptureStatus           MidtransStatus = `capture`
	MidtransSettlementStatus        MidtransStatus = `settlement`
	MidtransDenyStatus              MidtransStatus = `deny`
	MidtransCancelStatus            MidtransStatus = `cancel`
	MidtransExpireStatus            MidtransStatus = `expire`
	MidtransFailureStatus           MidtransStatus = `failure`
	MidtransRefundStatus            MidtransStatus = `refund`
	MidtransChargebackStatus        MidtransStatus = `chargeback`
	MidtransPartialRefundStatus     MidtransStatus = `partial_refund`
	MidtransPartialChargebackStatus MidtransStatus = `partial_chargeback`
	MidtransAuthorizedStatus        MidtransStatus = `authorized`
)
