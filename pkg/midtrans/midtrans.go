package midtrans_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type (
	MidtransConfig struct {
		ServerKey   string
		ClientKey   string
		Environment string
	}

	midtransClient struct {
		coreapi coreapi.Client
		snap    snap.Client
	}
)

func NewMidtransClient(config *MidtransConfig) (*midtransClient, error) {
	if config == nil || config.ServerKey == "" {
		return nil, errors.New("midtrans: server key is required")
	}
	env, err := midtransEnvironment(config.Environment)
	if err != nil {
		return nil, err
	}

	snapClient := snap.Client{}
	snapClient.New(config.ServerKey, env)
	coreClient := coreapi.Client{}
	coreClient.New(config.ServerKey, env)
	coreClient.ClientKey = config.ClientKey

	return &midtransClient{
		coreapi: coreClient,
		snap:    snapClient,
	}, nil
}

func midtransEnvironment(environment string) (midtrans.EnvironmentType, error) {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "", "sandbox":
		return midtrans.Sandbox, nil
	case "production":
		return midtrans.Production, nil
	default:
		return 0, fmt.Errorf("midtrans: unknown environment %q", environment)
	}
}

var _ MidtransClient = (*midtransClient)(nil)

func (c *midtransClient) CreateTransaction(ctx context.Context, request *TransactionRequest) (*TransactionResponse, error) {
	if err := contextReady(ctx); err != nil {
		return nil, err
	}
	if request == nil {
		return nil, errors.New("midtrans: transaction request is nil")
	}
	if request.TransactionDetails.OrderID == "" {
		return nil, errors.New("midtrans: transaction_details.order_id is required")
	}

	payload, err := asMap(request)
	if err != nil {
		return nil, err
	}
	body := snap.RequestParamWithMap(payload)
	responseMap, callErr := c.snapClient(ctx).CreateTransactionWithMap(&body)
	if err = sdkError(callErr); err != nil {
		return nil, err
	}

	response := &TransactionResponse{}
	if err = convert(responseMap, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) Charge(ctx context.Context, request *ChargeRequest) (*ChargeResponse, error) {
	if err := contextReady(ctx); err != nil {
		return nil, err
	}
	if request == nil {
		return nil, errors.New("midtrans: charge request is nil")
	}
	if request.PaymentType == "" {
		return nil, errors.New("midtrans: payment_type is required")
	}
	if request.TransactionDetails.OrderID == "" {
		return nil, errors.New("midtrans: transaction_details.order_id is required")
	}

	payload, err := asMap(request)
	if err != nil {
		return nil, err
	}
	body := coreapi.ChargeReqWithMap(payload)
	responseMap, callErr := c.coreClient(ctx).ChargeTransactionWithMap(&body)
	if err = sdkError(callErr); err != nil {
		return nil, err
	}

	response := &ChargeResponse{}
	if err = convert(responseMap, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) Capture(ctx context.Context, request *CaptureRequest) (*CaptureResponse, error) {
	if err := contextReady(ctx); err != nil {
		return nil, err
	}
	if request == nil {
		return nil, errors.New("midtrans: capture request is nil")
	}
	if request.TransactionID == "" {
		return nil, errors.New("midtrans: transaction_id is required")
	}

	captured, callErr := c.coreClient(ctx).CaptureTransaction(&coreapi.CaptureReq{
		TransactionID: request.TransactionID,
		GrossAmt:      float64(request.GrossAmount),
	})
	if err := sdkError(callErr); err != nil {
		return nil, err
	}

	response := &CaptureResponse{}
	if err := convert(captured, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) Cancel(ctx context.Context, request *CancelRequest) (*CancelResponse, error) {
	if err := contextReady(ctx); err != nil {
		return nil, err
	}
	if request == nil || request.OrderID == "" {
		return nil, errors.New("midtrans: order_id or transaction_id is required")
	}

	cancelled, callErr := c.coreClient(ctx).CancelTransaction(request.OrderID)
	if err := sdkError(callErr); err != nil {
		return nil, err
	}

	response := &CancelResponse{}
	if err := convert(cancelled, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *midtransClient) GetTransaction(ctx context.Context, transactionID string) (*HistoryTransactionResponse, error) {
	if err := contextReady(ctx); err != nil {
		return nil, err
	}
	if transactionID == "" {
		return nil, errors.New("midtrans: transaction_id is required")
	}
	transaction, callErr := c.coreClient(ctx).CheckTransaction(transactionID)
	if err := sdkError(callErr); err != nil {
		return nil, err
	}

	response := &HistoryTransactionResponse{}
	if err := convert(transaction, response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *midtransClient) snapClient(ctx context.Context) snap.Client {
	client := c.snap
	client.Options = optionsWithContext(client.Options, ctx)
	return client
}

func (c *midtransClient) coreClient(ctx context.Context) coreapi.Client {
	client := c.coreapi
	client.Options = optionsWithContext(client.Options, ctx)
	return client
}

func optionsWithContext(current *midtrans.ConfigOptions, ctx context.Context) *midtrans.ConfigOptions {
	options := &midtrans.ConfigOptions{}
	if current != nil {
		copied := *current
		options = &copied
	}
	if ctx != nil {
		options.SetContext(ctx)
	}
	return options
}

func contextReady(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func sdkError(err *midtrans.Error) error {
	if err == nil {
		return nil
	}
	return err
}

func asMap(value any) (map[string]any, error) {
	payload := map[string]any{}
	if err := convert(value, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func convert(src any, dst any) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}
