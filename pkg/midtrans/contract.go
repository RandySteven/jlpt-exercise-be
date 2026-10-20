package midtrans_client

// TransactionRequest is the Snap POST /snap/v1/transactions body.
// TransactionDetails is required.
type TransactionRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	ItemDetails        []ItemDetails      `json:"item_details,omitempty"`
	CustomerDetails    *CustomerDetails   `json:"customer_details,omitempty"`
	EnabledPayments    []string           `json:"enabled_payments,omitempty"`
	CreditCard         *CreditCard        `json:"credit_card,omitempty"`
	BCAVA              *BCAVirtualAccount `json:"bca_va,omitempty"`
	BNIVA              *VirtualAccount    `json:"bni_va,omitempty"`
	BRIVA              *VirtualAccount    `json:"bri_va,omitempty"`
	CIMBVA             *VirtualAccount    `json:"cimb_va,omitempty"`
	PermataVA          *PermataVA         `json:"permata_va,omitempty"`
	DanamonVA          *VirtualAccount    `json:"danamon_va,omitempty"`
	BSIVA              *VirtualAccount    `json:"bsi_va,omitempty"`
	SeaBankVA          *VirtualAccount    `json:"seabank_va,omitempty"`
	OtherVA            *VirtualAccount    `json:"other_va,omitempty"`
	Gopay              *Gopay             `json:"gopay,omitempty"`
	ShopeePay          *CallbackTarget    `json:"shopeepay,omitempty"`
	UOBEzpay           *CallbackTarget    `json:"uob_ezpay,omitempty"`
	Callbacks          *Callbacks         `json:"callbacks,omitempty"`
	Expiry             *Expiry            `json:"expiry,omitempty"`
	PageExpiry         *PageExpiry        `json:"page_expiry,omitempty"`
	Recurring          *Recurring         `json:"recurring,omitempty"`
	CustomField1       string             `json:"custom_field1,omitempty"`
	CustomField2       string             `json:"custom_field2,omitempty"`
	CustomField3       string             `json:"custom_field3,omitempty"`
}

// TransactionResponse is the Snap token response from POST /snap/v1/transactions.
type TransactionResponse struct {
	Token         string   `json:"token,omitempty"`
	RedirectURL   string   `json:"redirect_url,omitempty"`
	ErrorMessages []string `json:"error_messages,omitempty"`
	Locale        string   `json:"locale,omitempty"`
}

// TransactionDetails is the required transaction identity and amount.
type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

// ItemDetails is one line item. Price must not include a decimal.
type ItemDetails struct {
	ID           string `json:"id,omitempty"`
	Price        int64  `json:"price"`
	Quantity     int64  `json:"quantity"`
	Name         string `json:"name"`
	Brand        string `json:"brand,omitempty"`
	Category     string `json:"category,omitempty"`
	MerchantName string `json:"merchant_name,omitempty"`
	URL          string `json:"url,omitempty"`
}

// CustomerDetails is the payer shown on the payment page.
type CustomerDetails struct {
	FirstName       string   `json:"first_name,omitempty"`
	LastName        string   `json:"last_name,omitempty"`
	Email           string   `json:"email,omitempty"`
	Phone           string   `json:"phone,omitempty"`
	BillingAddress  *Address `json:"billing_address,omitempty"`
	ShippingAddress *Address `json:"shipping_address,omitempty"`
}

// Address is a billing or shipping address.
type Address struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

// CreditCard configures the Snap card payment method.
type CreditCard struct {
	SaveCard          bool               `json:"save_card,omitempty"`
	Secure            bool               `json:"secure,omitempty"`
	Channel           string             `json:"channel,omitempty"`
	Bank              string             `json:"bank,omitempty"`
	Type              string             `json:"type,omitempty"`
	Installment       *Installment       `json:"installment,omitempty"`
	WhitelistBins     []string           `json:"whitelist_bins,omitempty"`
	DynamicDescriptor *DynamicDescriptor `json:"dynamic_descriptor,omitempty"`
}

// Installment configures card installment options on Snap.
type Installment struct {
	Required bool              `json:"required,omitempty"`
	Disabled bool              `json:"disabled,omitempty"`
	Terms    *InstallmentTerms `json:"terms,omitempty"`
}

// InstallmentTerms lists available tenors per acquiring bank.
type InstallmentTerms struct {
	BNI     []int `json:"bni,omitempty"`
	Mandiri []int `json:"mandiri,omitempty"`
	CIMB    []int `json:"cimb,omitempty"`
	BCA     []int `json:"bca,omitempty"`
	BRI     []int `json:"bri,omitempty"`
	Mega    []int `json:"mega,omitempty"`
	Maybank []int `json:"maybank,omitempty"`
	Offline []int `json:"offline,omitempty"`
}

// DynamicDescriptor is the text printed on a BNI card statement.
type DynamicDescriptor struct {
	MerchantName string `json:"merchant_name,omitempty"`
	CityName     string `json:"city_name,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
}

// VirtualAccount is a custom VA number for banks that only take va_number.
type VirtualAccount struct {
	VANumber string `json:"va_number,omitempty"`
}

// BCAVirtualAccount is the BCA VA payment method.
type BCAVirtualAccount struct {
	VANumber       string    `json:"va_number,omitempty"`
	SubCompanyCode string    `json:"sub_company_code,omitempty"`
	FreeText       *FreeText `json:"free_text,omitempty"`
}

// PermataVA is the Permata VA payment method.
type PermataVA struct {
	VANumber      string `json:"va_number,omitempty"`
	RecipientName string `json:"recipient_name,omitempty"`
}

// FreeText is bilingual text shown during BCA VA inquiry and payment.
type FreeText struct {
	Inquiry []FreeTextItem `json:"inquiry,omitempty"`
	Payment []FreeTextItem `json:"payment,omitempty"`
}

// FreeTextItem is one bilingual line. Both languages are required when the item is sent.
type FreeTextItem struct {
	EN string `json:"en"`
	ID string `json:"id"`
}

// Gopay configures the Snap GoPay payment method.
type Gopay struct {
	EnableCallback      bool   `json:"enable_callback,omitempty"`
	CallbackURL         string `json:"callback_url,omitempty"`
	Tokenization        bool   `json:"tokenization,omitempty"`
	EnforceTokenization bool   `json:"enforce_tokenization,omitempty"`
	PhoneNumber         string `json:"phone_number,omitempty"`
	CountryCode         string `json:"country_code,omitempty"`
}

// CallbackTarget is a payment method that only sets a return URL.
type CallbackTarget struct {
	CallbackURL string `json:"callback_url,omitempty"`
}

// Callbacks are Snap redirect URLs after the customer finishes or fails payment.
type Callbacks struct {
	Finish string `json:"finish,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Expiry overrides how long the payment stays payable.
// StartTime uses yyyy-MM-dd HH:mm:ss Z. Unit is day, hour, or minute.
type Expiry struct {
	StartTime string `json:"start_time,omitempty"`
	Unit      string `json:"unit"`
	Duration  int64  `json:"duration"`
}

// PageExpiry overrides how long the Snap page stays open.
type PageExpiry struct {
	Duration int64  `json:"duration"`
	Unit     string `json:"unit"`
}

// Recurring turns on the Snap subscription UI.
// IntervalUnit is the subscription charge unit, for example week.
type Recurring struct {
	Required     bool   `json:"required"`
	StartTime    string `json:"start_time,omitempty"`
	IntervalUnit string `json:"interval_unit,omitempty"`
}

// ChargeRequest is the Core API POST /v2/charge body.
type ChargeRequest struct {
	PaymentType        string              `json:"payment_type"`
	TransactionDetails TransactionDetails  `json:"transaction_details"`
	ItemDetails        []ItemDetails       `json:"item_details,omitempty"`
	CustomerDetails    *CustomerDetails    `json:"customer_details,omitempty"`
	CreditCard         *ChargeCreditCard   `json:"credit_card,omitempty"`
	BankTransfer       *BankTransfer       `json:"bank_transfer,omitempty"`
	EChannel           *EChannel           `json:"echannel,omitempty"`
	Gopay              *ChargeGopay        `json:"gopay,omitempty"`
	ShopeePay          *CallbackTarget     `json:"shopeepay,omitempty"`
	Qris               *Qris               `json:"qris,omitempty"`
	ConvenienceStore   *ConvenienceStore   `json:"cstore,omitempty"`
	CustomExpiry       *ChargeCustomExpiry `json:"custom_expiry,omitempty"`
	CustomField1       string              `json:"custom_field1,omitempty"`
	CustomField2       string              `json:"custom_field2,omitempty"`
	CustomField3       string              `json:"custom_field3,omitempty"`
}

// ChargeCreditCard is the card object for Core API charge.
// TokenID comes from GET /v2/token.
type ChargeCreditCard struct {
	TokenID           string   `json:"token_id"`
	Authentication    bool     `json:"authentication,omitempty"`
	Bank              string   `json:"bank,omitempty"`
	InstallmentTerm   int      `json:"installment_term,omitempty"`
	Type              string   `json:"type,omitempty"`
	Bins              []string `json:"bins,omitempty"`
	SaveTokenID       bool     `json:"save_token_id,omitempty"`
	PointRedeemAmount int64    `json:"point_redeem_amount,omitempty"`
	CallbackType      string   `json:"callback_type,omitempty"`
}

// BankTransfer is the Core API bank_transfer charge object.
type BankTransfer struct {
	Bank     string             `json:"bank"`
	VANumber string             `json:"va_number,omitempty"`
	Permata  *PermataVA         `json:"permata,omitempty"`
	FreeText *FreeText          `json:"free_text,omitempty"`
	BCA      *BCATransferDetail `json:"bca,omitempty"`
}

// BCATransferDetail sets the BCA sub company code on a bank transfer charge.
type BCATransferDetail struct {
	SubCompanyCode string `json:"sub_company_code,omitempty"`
}

// EChannel is the Mandiri Bill charge object.
type EChannel struct {
	BillInfo1 string `json:"bill_info1"`
	BillInfo2 string `json:"bill_info2"`
	BillInfo3 string `json:"bill_info3,omitempty"`
	BillInfo4 string `json:"bill_info4,omitempty"`
	BillInfo5 string `json:"bill_info5,omitempty"`
	BillInfo6 string `json:"bill_info6,omitempty"`
	BillInfo7 string `json:"bill_info7,omitempty"`
	BillInfo8 string `json:"bill_info8,omitempty"`
	BillKey   string `json:"bill_key,omitempty"`
}

// ChargeGopay is the Core API GoPay charge object.
type ChargeGopay struct {
	EnableCallback     bool   `json:"enable_callback,omitempty"`
	CallbackURL        string `json:"callback_url,omitempty"`
	AccountID          string `json:"account_id,omitempty"`
	PaymentOptionToken string `json:"payment_option_token,omitempty"`
	PreAuth            bool   `json:"pre_auth,omitempty"`
	Recurring          bool   `json:"recurring,omitempty"`
}

// Qris selects the QRIS acquirer for a Core API charge.
type Qris struct {
	Acquirer string `json:"acquirer,omitempty"`
}

// ConvenienceStore is the Core API over-the-counter charge object.
type ConvenienceStore struct {
	Store   string `json:"store"`
	Message string `json:"message,omitempty"`
}

// ChargeCustomExpiry sets how long a Core API charge stays payable.
// OrderTime uses yyyy-MM-dd HH:mm:ss Z. Unit is second, minute, hour, or day.
type ChargeCustomExpiry struct {
	OrderTime      string `json:"order_time,omitempty"`
	ExpiryDuration int64  `json:"expiry_duration,omitempty"`
	Unit           string `json:"unit,omitempty"`
}

// CaptureRequest captures an authorized card or GoPay transaction.
// GrossAmount is optional and captures the full amount when omitted.
type CaptureRequest struct {
	TransactionID string `json:"transaction_id"`
	GrossAmount   int64  `json:"gross_amount,omitempty"`
}

// CancelRequest voids a transaction that is not yet settled.
// OrderID is the merchant order id or the Midtrans transaction id.
type CancelRequest struct {
	OrderID string `json:"-"`
}

// ChargeResponse is the Core API charge result.
type ChargeResponse struct {
	CoreTransactionResponse
}

// CaptureResponse is the Core API capture result.
type CaptureResponse struct {
	CoreTransactionResponse
}

// CancelResponse is the Core API cancel result.
type CancelResponse struct {
	CoreTransactionResponse
}

// CoreTransactionResponse is the shared Core API transaction payload.
type CoreTransactionResponse struct {
	StatusCode             string     `json:"status_code,omitempty"`
	StatusMessage          string     `json:"status_message,omitempty"`
	TransactionID          string     `json:"transaction_id,omitempty"`
	OrderID                string     `json:"order_id,omitempty"`
	GrossAmount            string     `json:"gross_amount,omitempty"`
	PaymentType            string     `json:"payment_type,omitempty"`
	TransactionTime        string     `json:"transaction_time,omitempty"`
	TransactionStatus      string     `json:"transaction_status,omitempty"`
	FraudStatus            string     `json:"fraud_status,omitempty"`
	MaskedCard             string     `json:"masked_card,omitempty"`
	Bank                   string     `json:"bank,omitempty"`
	ApprovalCode           string     `json:"approval_code,omitempty"`
	ECI                    string     `json:"eci,omitempty"`
	SignatureKey           string     `json:"signature_key,omitempty"`
	Currency               string     `json:"currency,omitempty"`
	ChannelResponseCode    string     `json:"channel_response_code,omitempty"`
	ChannelResponseMessage string     `json:"channel_response_message,omitempty"`
	RedirectURL            string     `json:"redirect_url,omitempty"`
	PermataVANumber        string     `json:"permata_va_number,omitempty"`
	VANumbers              []VANumber `json:"va_numbers,omitempty"`
	BillKey                string     `json:"bill_key,omitempty"`
	BillerCode             string     `json:"biller_code,omitempty"`
	Actions                []Action   `json:"actions,omitempty"`
	PaymentCode            string     `json:"payment_code,omitempty"`
	QRString               string     `json:"qr_string,omitempty"`
	ExpiryTime             string     `json:"expiry_time,omitempty"`
}

// VANumber is a virtual account returned by a charge.
type VANumber struct {
	Bank     string `json:"bank,omitempty"`
	VANumber string `json:"va_number,omitempty"`
}

// Action is a next step returned by a charge, such as a deeplink or QR URL.
type Action struct {
	Name   string   `json:"name,omitempty"`
	Method string   `json:"method,omitempty"`
	URL    string   `json:"url,omitempty"`
	Fields []string `json:"fields,omitempty"`
}
