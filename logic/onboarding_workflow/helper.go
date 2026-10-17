package onboarding_workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/golang-jwt/jwt/v5"
)

func onboardingRegisterRequestMapping(request *requests.RegisterRequest) *models.Onboarding {
	password, err := utils.HashPassword(request.Password)
	if err != nil {
		log.Println("failed to hash password", err)
		return nil
	}
	additionalInfo := request.AdditionalInfo
	if len(additionalInfo) == 0 {
		additionalInfo = json.RawMessage(`{}`)
	}
	return &models.Onboarding{
		Name:           mergeName(request.FirstName, request.LastName),
		Email:          request.Email,
		Password:       password,
		PhoneNumber:    request.PhoneNumber,
		Address:        request.Address,
		RegisterAs:     enums.RegisterAsUser,
		Status:         `PROCESSING`,
		AdditionalInfo: additionalInfo,
	}
}

func mappingUserInfo(additionalInfo json.RawMessage) *models.User {
	if len(additionalInfo) == 0 {
		additionalInfo = json.RawMessage(`{}`)
	}
	return &models.User{
		AdditionalInfo: additionalInfo,
	}
}

func mergeName(firstName, lastName string) string {
	if lastName == "" {
		return firstName
	}
	return fmt.Sprintf("%s %s", firstName, lastName)
}

var secretKey = []byte("activation_hash")

type DecodeActivationTokenObject struct {
	jwt.RegisteredClaims
	UserID         uint64    `json:"user_id"`
	OnboardingID   uint64    `json:"onboarding_id"`
	WorkflowID     string    `json:"workflow_id"`
	RunID          string    `json:"run_id"`
	ExpirationTime time.Time `json:"expiration_time"`
}

func activationTokenGenerate(userID, onboardingID uint64, workflowID string, runID string) string {
	claims := &DecodeActivationTokenObject{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ActivationRequest",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
		UserID:       userID,
		OnboardingID: onboardingID,
		WorkflowID:   workflowID,
		RunID:        runID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return ""
	}
	return signedToken
}

func decodeActivationToken(tokenString string) (*DecodeActivationTokenObject, error) {
	claims := &DecodeActivationTokenObject{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func registerResponseReady(resp *responses.RegisterResponse) bool {
	return resp != nil && (resp.ActivationToken != "" || resp.OnboardingID != 0)
}
