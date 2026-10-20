package midtrans_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type midtransClient struct {
	client midtrans.HttpClient
	snap   snap.Client
}

func NewMidtransClient(client midtrans.HttpClient, snap snap.Client) *midtransClient {
	return &midtransClient{
		client: client,
		snap:   snap,
	}
}

var _ MidtransClient = (*midtransClient)(nil)

func (c *midtransClient) CreateTransaction(ctx context.Context, request *TransactionRequest) (*TransactionResponse, error) {
	if request == nil {
		return nil, errors.New("midtrans: transaction request is nil")
	}
	if request.TransactionDetails.OrderID == "" {
		return nil, errors.New("midtrans: transaction_details.order_id is required")
	}

	response := &TransactionResponse{}
	endpoint := c.snap.Env.SnapURL() + "/snap/v1/transactions"
	if err := c.call(ctx, http.MethodPost, endpoint, request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) Charge(ctx context.Context, request *ChargeRequest) (*ChargeResponse, error) {
	if request == nil {
		return nil, errors.New("midtrans: charge request is nil")
	}
	if request.PaymentType == "" {
		return nil, errors.New("midtrans: payment_type is required")
	}
	if request.TransactionDetails.OrderID == "" {
		return nil, errors.New("midtrans: transaction_details.order_id is required")
	}

	response := &ChargeResponse{}
	endpoint := c.snap.Env.BaseUrl() + "/v2/charge"
	if err := c.call(ctx, http.MethodPost, endpoint, request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) Capture(ctx context.Context, request *CaptureRequest) (*CaptureResponse, error) {
	if request == nil {
		return nil, errors.New("midtrans: capture request is nil")
	}
	if request.TransactionID == "" {
		return nil, errors.New("midtrans: transaction_id is required")
	}

	response := &CaptureResponse{}
	endpoint := c.snap.Env.BaseUrl() + "/v2/capture"
	if err := c.call(ctx, http.MethodPost, endpoint, request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) Cancel(ctx context.Context, request *CancelRequest) (*CancelResponse, error) {
	if request == nil || request.OrderID == "" {
		return nil, errors.New("midtrans: order_id or transaction_id is required")
	}

	response := &CancelResponse{}
	endpoint := fmt.Sprintf("%s/v2/%s/cancel", c.snap.Env.BaseUrl(), url.PathEscape(request.OrderID))
	if err := c.call(ctx, http.MethodPost, endpoint, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) call(ctx context.Context, method, endpoint string, payload any, result any) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	options := &midtrans.ConfigOptions{}
	if ctx != nil {
		options.SetContext(ctx)
	}
	if c.snap.Options != nil {
		options.PaymentIdempotencyKey = c.snap.Options.PaymentIdempotencyKey
		options.PaymentOverrideNotification = c.snap.Options.PaymentOverrideNotification
		options.PaymentAppendNotification = c.snap.Options.PaymentAppendNotification
	}

	serverKey := c.snap.ServerKey
	if err := c.client.Call(method, endpoint, &serverKey, options, body, result); err != nil {
		return err
	}
	return nil
}
