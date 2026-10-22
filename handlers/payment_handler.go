package handlers

import "net/http"

type (
	IPaymentHandler interface {
		CreateSubscription(w http.ResponseWriter, r *http.Request)
		PaymentConfirmation(w http.ResponseWriter, r *http.Request)
		CreateCharge(w http.ResponseWriter, r *http.Request)
		CaptureCharge(w http.ResponseWriter, r *http.Request)
		CancelCharge(w http.ResponseWriter, r *http.Request)
		RefundCharge(w http.ResponseWriter, r *http.Request)
		GetCharge(w http.ResponseWriter, r *http.Request)
		GetChargeList(w http.ResponseWriter, r *http.Request)
		GetChargeDetail(w http.ResponseWriter, r *http.Request)
		GetChargeTransaction(w http.ResponseWriter, r *http.Request)
		GetChargeTransactionList(w http.ResponseWriter, r *http.Request)
		GetChargeTransactionDetail(w http.ResponseWriter, r *http.Request)
	}

	PaymentHandler struct{}
)

func (p *PaymentHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) PaymentConfirmation(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) CreateCharge(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) CaptureCharge(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) CancelCharge(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) RefundCharge(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) GetCharge(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) GetChargeList(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) GetChargeDetail(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) GetChargeTransaction(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) GetChargeTransactionList(w http.ResponseWriter, r *http.Request) {

}

func (p *PaymentHandler) GetChargeTransactionDetail(w http.ResponseWriter, r *http.Request) {

}