package payment_workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/jackc/pgx/v5/pgconn"
	"go.temporal.io/api/serviceerror"
)

const (
	paymentStatusPending = "PENDING"
	paymentStatusPaid    = "PAID"
	paymentStatusFailed  = "FAILED"

	subscriptionStatusPending = "PENDING"
	subscriptionStatusActive  = "ACTIVE"
	subscriptionStatusFailed  = "FAILED"

	paymentInputErrorType    = "PaymentInputError"
	paymentRejectedErrorType = "PaymentRejected"
)

type paymentMetadata struct {
	PlanID      uint64 `json:"plan_id"`
	PlanName    string `json:"plan_name"`
	Token       string `json:"token,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

func paymentWorkflowID(idempotencyKey string) string {
	return fmt.Sprintf("%s-%s", paymentWorkflowName, idempotencyKey)
}

func (e *ExecutionData) setSubscriptionResponse(id uint64) {
	if e == nil || id == 0 {
		return
	}
	if e.Response == nil {
		e.Response = &responses.CreateSubscriptionResponse{}
	}
	e.Response.SubscriptionID = strconv.FormatUint(id, 10)
}

func subscriptionEnd(start time.Time, plan *models.SubscriptionPlan) (time.Time, error) {
	if plan == nil {
		return time.Time{}, fmt.Errorf("plan is missing")
	}
	count := plan.IntervalCount
	if count <= 0 {
		return time.Time{}, fmt.Errorf("plan interval is invalid")
	}
	switch strings.ToLower(strings.TrimSpace(plan.Interval)) {
	case "day", "days":
		return start.AddDate(0, 0, count), nil
	case "week", "weeks":
		return start.AddDate(0, 0, 7*count), nil
	case "month", "months":
		return start.AddDate(0, count, 0), nil
	case "year", "years":
		return start.AddDate(count, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("plan interval %q is invalid", plan.Interval)
	}
}

func encodePaymentMetadata(meta paymentMetadata) (json.RawMessage, error) {
	raw, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func decodePaymentMetadata(raw json.RawMessage) paymentMetadata {
	meta := paymentMetadata{}
	if len(raw) == 0 {
		return meta
	}
	_ = json.Unmarshal(raw, &meta)
	return meta
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func alreadyStarted(err error) bool {
	var target *serviceerror.WorkflowExecutionAlreadyStarted
	return errors.As(err, &target)
}

func rejectedMidtrans(err error) bool {
	return err != nil && strings.Contains(err.Error(), "400")
}
