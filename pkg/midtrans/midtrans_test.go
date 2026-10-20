package midtrans_client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type fakeHTTP struct {
	method  string
	url     string
	apiKey  string
	body    []byte
	hasBody bool
	result  any
	err     *midtrans.Error
}

func (f *fakeHTTP) Call(method string, url string, apiKey *string, _ *midtrans.ConfigOptions, body io.Reader, result interface{}) *midtrans.Error {
	f.method = method
	f.url = url
	if apiKey != nil {
		f.apiKey = *apiKey
	}
	if body != nil {
		raw, err := io.ReadAll(body)
		if err != nil {
			return &midtrans.Error{Message: err.Error(), RawError: err}
		}
		f.body = raw
		f.hasBody = true
	}
	if f.err != nil {
		return f.err
	}
	if f.result != nil && result != nil {
		raw, err := json.Marshal(f.result)
		if err != nil {
			return &midtrans.Error{Message: err.Error(), RawError: err}
		}
		if err := json.Unmarshal(raw, result); err != nil {
			return &midtrans.Error{Message: err.Error(), RawError: err}
		}
	}
	return nil
}

func testClient(fake *fakeHTTP) *midtransClient {
	return NewMidtransClient(fake, snap.Client{
		ServerKey: "SB-Mid-server-test",
		Env:       midtrans.Sandbox,
	})
}

func TestCreateTransactionUsesSnapBackendContract(t *testing.T) {
	fake := &fakeHTTP{result: TransactionResponse{
		Token:       "snap-token",
		RedirectURL: "https://app.sandbox.midtrans.com/snap/v3/redirection/snap-token",
	}}
	client := testClient(fake)

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
	if fake.method != http.MethodPost {
		t.Fatalf("method = %s", fake.method)
	}
	if fake.url != "https://app.sandbox.midtrans.com/snap/v1/transactions" {
		t.Fatalf("url = %s", fake.url)
	}
	if fake.apiKey != "SB-Mid-server-test" {
		t.Fatalf("api key = %s", fake.apiKey)
	}
	if response.Token != "snap-token" || response.RedirectURL == "" {
		t.Fatalf("response = %+v", response)
	}

	var payload map[string]any
	if err := json.Unmarshal(fake.body, &payload); err != nil {
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
		fake := &fakeHTTP{result: ChargeResponse{CoreTransactionResponse: CoreTransactionResponse{
			StatusCode: "201", TransactionStatus: "pending", OrderID: "ORDER-101",
		}}}
		client := testClient(fake)
		response, err := client.Charge(context.Background(), &ChargeRequest{
			PaymentType:        "bank_transfer",
			TransactionDetails: TransactionDetails{OrderID: "ORDER-101", GrossAmount: 10000},
			BankTransfer:       &BankTransfer{Bank: "bca"},
		})
		if err != nil {
			t.Fatalf("charge: %v", err)
		}
		if fake.url != "https://api.sandbox.midtrans.com/v2/charge" {
			t.Fatalf("url = %s", fake.url)
		}
		if response.TransactionStatus != "pending" {
			t.Fatalf("status = %s", response.TransactionStatus)
		}
	})

	t.Run("capture", func(t *testing.T) {
		fake := &fakeHTTP{result: CaptureResponse{CoreTransactionResponse: CoreTransactionResponse{
			StatusCode: "200", TransactionStatus: "capture", GrossAmount: "30000.00",
		}}}
		client := testClient(fake)
		response, err := client.Capture(context.Background(), &CaptureRequest{
			TransactionID: "be4f3e44-d6ee-4355-8c64-c1d1dc7f4590",
			GrossAmount:   145000,
		})
		if err != nil {
			t.Fatalf("capture: %v", err)
		}
		if fake.url != "https://api.sandbox.midtrans.com/v2/capture" {
			t.Fatalf("url = %s", fake.url)
		}
		var payload map[string]any
		if err := json.Unmarshal(fake.body, &payload); err != nil {
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
		fake := &fakeHTTP{result: CancelResponse{CoreTransactionResponse: CoreTransactionResponse{
			StatusCode: "200", StatusMessage: "Success, transaction is canceled", TransactionStatus: "cancel",
		}}}
		client := testClient(fake)
		response, err := client.Cancel(context.Background(), &CancelRequest{OrderID: "example-1424936368"})
		if err != nil {
			t.Fatalf("cancel: %v", err)
		}
		if fake.url != "https://api.sandbox.midtrans.com/v2/example-1424936368/cancel" {
			t.Fatalf("url = %s", fake.url)
		}
		if fake.hasBody {
			t.Fatalf("cancel body = %s", fake.body)
		}
		if response.TransactionStatus != "cancel" {
			t.Fatalf("status = %s", response.TransactionStatus)
		}
	})
}

func TestClientRejectsMissingRequiredFields(t *testing.T) {
	client := testClient(&fakeHTTP{})
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

func TestClientReturnsAPIError(t *testing.T) {
	fake := &fakeHTTP{err: &midtrans.Error{Message: "transaction_details.order_id is required", StatusCode: 400}}
	client := testClient(fake)
	_, err := client.CreateTransaction(context.Background(), &TransactionRequest{
		TransactionDetails: TransactionDetails{OrderID: "ORDER-101", GrossAmount: 10000},
	})
	if err == nil {
		t.Fatal("expected api error")
	}
}
