package onboarding_workflow

import (
	"strings"
	"testing"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/golang-jwt/jwt/v5"
)

func TestOnboardingRegisterRequestMapping(t *testing.T) {
	if onboardingRegisterRequestMapping(&requests.RegisterRequest{Password: strings.Repeat("a", 73)}) != nil {
		t.Fatal("oversized password was mapped")
	}

	got := onboardingRegisterRequestMapping(&requests.RegisterRequest{
		FirstName:      "Ada",
		Email:          "ada@example.com",
		Password:       "secret",
		PhoneNumber:    "0800",
		Address:        "London",
		AdditionalInfo: []byte(`{"city":"London"}`),
	})
	if got == nil {
		t.Fatal("mapping returned nil")
	}
	if got.Name != "Ada" || got.Email != "ada@example.com" || got.PhoneNumber != "0800" || got.Address != "London" || got.RegisterAs != enums.RegisterAsUser || got.Status != "PROCESSING" {
		t.Fatalf("mapped = %#v", got)
	}
	if string(got.AdditionalInfo) != `{"city":"London"}` || !utils.ComparePassword("secret", got.Password) {
		t.Fatalf("additional info %s password hashed = %v", got.AdditionalInfo, got.Password != "secret")
	}

	empty := onboardingRegisterRequestMapping(&requests.RegisterRequest{FirstName: "Ada", LastName: "Lovelace", Password: "secret"})
	if empty.Name != "Ada Lovelace" || string(empty.AdditionalInfo) != "{}" {
		t.Fatalf("empty additional info = %#v", empty)
	}
}

func TestMappingUserInfoUsesEmptyObject(t *testing.T) {
	got := mappingUserInfo(nil)
	if string(got.AdditionalInfo) != "{}" {
		t.Fatalf("additional info = %s", got.AdditionalInfo)
	}
}

func TestMergeName(t *testing.T) {
	if mergeName("Ada", "") != "Ada" {
		t.Fatal("empty last name was not dropped")
	}
	if mergeName("Ada", "Lovelace") != "Ada Lovelace" {
		t.Fatal("full name was not joined")
	}
}

func TestActivationToken(t *testing.T) {
	token := activationTokenGenerate(4, 9, "wf", "run")
	claims, err := decodeActivationToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 4 || claims.OnboardingID != 9 || claims.WorkflowID != "wf" || claims.RunID != "run" || claims.ExpiresAt == nil {
		t.Fatalf("claims = %#v", claims)
	}
	if !claims.ExpiresAt.After(time.Now()) {
		t.Fatal("token was already expired")
	}

	if _, err = decodeActivationToken("not-a-token"); err == nil {
		t.Fatal("malformed token was accepted")
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, &DecodeActivationTokenObject{UserID: 1})
	signed, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = decodeActivationToken(signed); err == nil || !strings.Contains(err.Error(), "unexpected signing method") {
		t.Fatalf("none algorithm: %v", err)
	}
}

func TestRegisterResponseReady(t *testing.T) {
	if registerResponseReady(nil) || registerResponseReady(&responses.RegisterResponse{}) {
		t.Fatal("empty response was ready")
	}
	if !registerResponseReady(&responses.RegisterResponse{ActivationToken: "token"}) {
		t.Fatal("token response was not ready")
	}
	if !registerResponseReady(&responses.RegisterResponse{OnboardingID: 3}) {
		t.Fatal("onboarding response was not ready")
	}
}
