package midtrans_client

import "context"

type MidtransClient interface {
	CreateTransaction(ctx context.Context, request *TransactionRequest) (*TransactionResponse, error)
	Charge(ctx context.Context, request *ChargeRequest) (*ChargeResponse, error)
	Capture(ctx context.Context, request *CaptureRequest) (*CaptureResponse, error)
	Cancel(ctx context.Context, request *CancelRequest) (*CancelResponse, error)
}
