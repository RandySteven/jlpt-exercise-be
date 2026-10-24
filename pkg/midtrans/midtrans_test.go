package midtrans_client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	midtransmocks "github.com/RandySteven/jlpt-exercise-be/mocks/midtrans"
	"github.com/midtrans/midtrans-go"
	"github.com/stretchr/testify/mock"
)

type capturedCall struct {
	method  string
	url     string
	apiKey  string
	body    []byte
	hasBody bool
}

func testClient(t *testing.T, result any, callErr *midtrans.Error) (*midtransClient, *capturedCall) {
	t.Helper()
	httpClient := midtransmocks.NewMockHttpClient(t)
	captured := &capturedCall{}
	httpClient.EXPECT().Call(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(method string, url string, apiKey *string, _ *midtrans.ConfigOptions, body io.Reader, dest interface{}) *midtrans.Error {
			captured.method = method
			captured.url = url
			if apiKey != nil {
				captured.apiKey = *apiKey
			}
			if body != nil {
				raw, err := io.ReadAll(body)
				if err != nil {
					return &midtrans.Error{Message: err.Error(), RawError: err}
				}
				captured.body = raw
				captured.hasBody = true
			}
			if callErr != nil {
				return callErr
			}
			if result != nil && dest != nil {
				raw, err := json.Marshal(result)
				if err != nil {
					return &midtrans.Error{Message: err.Error(), RawError: err}
				}
				if err := json.Unmarshal(raw, dest); err != nil {
					return &midtrans.Error{Message: err.Error(), RawError: err}
				}
			}
			return nil
		}).Maybe()

	client, err := NewMidtransClient(&MidtransConfig{
		ServerKey:   "SB-Mid-server-test",
		Environment: "sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}
	client.coreapi.HttpClient = httpClient
	client.snap.HttpClient = httpClient
	return client, captured
}

func TestCreateTransactionUsesSnapBackendContract(t *testing.T) {
	client, captured := testClient(t, TransactionResponse{
		Token:       "snap-token",
		RedirectURL: "https://app.sandbox.midtrans.com/snap/v3/redirection/snap-token",
	}, nil)

	response, err := client.CreateTransaction(context.Background(), &TransactionRequest{
		TransactionDetails: TransactionDetails{OrderID: "ORDER-101", GrossAmount: 10000},
		ItemDetails: []ItemDetails{{
			ID: "ITEM1", Price: 10000, Quantity: 1, Name: "Midtrans Bear",
		}},
		EnabledPayments: []string{"credit_card", "gopay"},
		CreditCard:      &CreditCard{Secure: true},
		Callbacks:       &Callbacks{Finish: "https://demo.midtrans.com"},
		PageExpiry:      &PageExpiry{Duration: 3, Unit: "hours"},
		Recurring:       &Recurring{Required: true, IntervalUnit: "week"},
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if captured.method != http.MethodPost {
		t.Fatalf("method = %s", captured.method)
	}
	if captured.url != "https://app.sandbox.midtrans.com/snap/v1/transactions" {
		t.Fatalf("url = %s", captured.url)
	}
	if captured.apiKey != "SB-Mid-server-test" {
		t.Fatalf("api key = %s", captured.apiKey)
	}
	if response.Token != "snap-token" || response.RedirectURL == "" {
		t.Fatalf("response = %+v", response)
	}

	var payload map[string]any
	if err := json.Unmarshal(captured.body, &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	details := payload["transaction_details"].(map[string]any)
	if details["order_id"] != "ORDER-101" || details["gross_amount"] != float64(10000) {
		t.Fatalf("transaction_details = %#v", details)
	}
	if _, ok := payload["page_expiry"]; !ok {
		t.Fatal("page_expiry missing")
	}
	if _, ok := payload["recurring"]; !ok {
		t.Fatal("recurring missing")
	}
}

func TestChargeCaptureAndCancelUseCoreAPI(t *testing.T) {
	t.Run("charge", func(t *testing.T) {
		client, captured := testClient(t, ChargeResponse{CoreTransactionResponse: CoreTransactionResponse{
			StatusCode: "201", TransactionStatus: "pending", OrderID: "ORDER-101",
		}}, nil)
		response, err := client.Charge(context.Background(), &ChargeRequest{
			PaymentType:        "bank_transfer",
			TransactionDetails: TransactionDetails{OrderID: "ORDER-101", GrossAmount: 10000},
			BankTransfer:       &BankTransfer{Bank: "bca"},
		})
		if err != nil {
			t.Fatalf("charge: %v", err)
		}
		if captured.url != "https://api.sandbox.midtrans.com/v2/charge" {
			t.Fatalf("url = %s", captured.url)
		}
		if response.TransactionStatus != "pending" {
			t.Fatalf("status = %s", response.TransactionStatus)
		}
	})

	t.Run("capture", func(t *testing.T) {
		client, captured := testClient(t, CaptureResponse{CoreTransactionResponse: CoreTransactionResponse{
			StatusCode: "200", TransactionStatus: "capture", GrossAmount: "30000.00",
		}}, nil)
		response, err := client.Capture(context.Background(), &CaptureRequest{
			TransactionID: "be4f3e44-d6ee-4355-8c64-c1d1dc7f4590",
			GrossAmount:   145000,
		})
		if err != nil {
			t.Fatalf("capture: %v", err)
		}
		if captured.url != "https://api.sandbox.midtrans.com/v2/capture" {
			t.Fatalf("url = %s", captured.url)
		}
		var payload map[string]any
		if err := json.Unmarshal(captured.body, &payload); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if payload["transaction_id"] != "be4f3e44-d6ee-4355-8c64-c1d1dc7f4590" || payload["gross_amount"] != float64(145000) {
			t.Fatalf("payload = %#v", payload)
		}
		if response.StatusCode != "200" || response.GrossAmount != "30000.00" {
			t.Fatalf("response = %+v", response)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		client, captured := testClient(t, CancelResponse{CoreTransactionResponse: CoreTransactionResponse{
			StatusCode: "200", StatusMessage: "Success, transaction is canceled", TransactionStatus: "cancel",
		}}, nil)
		response, err := client.Cancel(context.Background(), &CancelRequest{OrderID: "example-1424936368"})
		if err != nil {
			t.Fatalf("cancel: %v", err)
		}
		if captured.url != "https://api.sandbox.midtrans.com/v2/example-1424936368/cancel" {
			t.Fatalf("url = %s", captured.url)
		}
		if captured.hasBody {
			t.Fatalf("cancel body = %s", captured.body)
		}
		if response.TransactionStatus != "cancel" {
			t.Fatalf("status = %s", response.TransactionStatus)
		}
	})
}

func TestClientRejectsMissingRequiredFields(t *testing.T) {
	client, _ := testClient(t, nil, nil)
	if _, err := client.CreateTransaction(context.Background(), &TransactionRequest{}); err == nil {
		t.Fatal("expected missing order id error")
	}
	if _, err := client.Charge(context.Background(), &ChargeRequest{}); err == nil {
		t.Fatal("expected missing payment type error")
	}
	if _, err := client.Capture(context.Background(), &CaptureRequest{}); err == nil {
		t.Fatal("expected missing transaction id error")
	}
	if _, err := client.Cancel(context.Background(), &CancelRequest{}); err == nil {
		t.Fatal("expected missing order id error")
	}
}

func TestNewMidtransClient(t *testing.T) {
	if _, err := NewMidtransClient(nil); err == nil {
		t.Fatal("expected missing config error")
	}
	if _, err := NewMidtransClient(&MidtransConfig{Environment: "sandbox"}); err == nil {
		t.Fatal("expected missing server key error")
	}
	if _, err := NewMidtransClient(&MidtransConfig{ServerKey: "server-key", Environment: "staging"}); err == nil {
		t.Fatal("expected unknown environment error")
	}
	client, err := NewMidtransClient(&MidtransConfig{ServerKey: "server-key", ClientKey: "client-key", Environment: "sandbox"})
	if err != nil || client == nil || client.coreapi.ServerKey != "server-key" || client.coreapi.ClientKey != "client-key" {
		t.Fatalf("client = %#v err %v", client, err)
	}
}

func TestClientReturnsAPIError(t *testing.T) {
	client, _ := testClient(t, nil, &midtrans.Error{Message: "transaction_details.order_id is required", StatusCode: 400})
	_, err := client.CreateTransaction(context.Background(), &TransactionRequest{
		TransactionDetails: TransactionDetails{OrderID: "ORDER-101", GrossAmount: 10000},
	})
	if err == nil {
		t.Fatal("expected api error")
	}
}
