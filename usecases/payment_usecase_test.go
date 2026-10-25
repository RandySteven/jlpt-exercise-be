package usecases

import (
	"testing"

	cachemocks "github.com/RandySteven/jlpt-exercise-be/caches/mocks"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/stretchr/testify/mock"
)

func TestGetUserSubscriptionDetailReturnsCachedDetail(t *testing.T) {
	cache := cachemocks.NewMockPaymentCache(t)
	cache.EXPECT().GetUserSubscriptionDetail(mock.Anything, "7").Return(&responses.GetUserSubscriptionDetailResponse{
		SubscriptionID: 9,
		PlanName:       "N3",
	}, nil)

	usecase := newPaymentUsecase(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, cache)

	got, customErr := usecase.GetUserSubscriptionDetail(userContext(7))
	if customErr != nil {
		t.Fatalf("error = %#v", customErr)
	}
	if got == nil || got.SubscriptionID != 9 || got.PlanName != "N3" {
		t.Fatalf("response = %#v", got)
	}
}
