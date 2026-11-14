package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

func UserFullName(firstName, lastName string) string {
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func StringToDate(dateStr string) (time.Time, error) {
	dateTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return dateTime, nil
}

func HashImageFile(imageFile string) string {
	return imageFile
}

func GetFieldsOfObject(object interface{}) []string {
	fields := []string{}
	typ := reflect.TypeOf(object).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i).Name
		fields = append(fields, field)
	}
	return fields
}

func ReplaceLastURLID(url string) string {
	return strings.Replace(url, "/{id}", "", -1)
}

func WriteLogFile() (*os.File, error) {
	year, month, day := time.Now().Date()
	monthStr, dayStr := fmt.Sprintf("%d", month), fmt.Sprintf("%d", day)
	if month < 10 {
		monthStr = "0" + monthStr
	}
	if day < 10 {
		dayStr = "0" + dayStr
	}
	dateFile := fmt.Sprintf("%d%s%s.log", year, monthStr, dayStr)

	logFile, err := os.OpenFile(dateFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	if err != nil {
		return nil, err
	}

	return logFile, nil
}

func FeatureFlagsToMap(featureFlags json.RawMessage) map[string]bool {
	featureFlagsMap := make(map[string]bool)
	err := json.Unmarshal(featureFlags, &featureFlagsMap)
	if err != nil {
		return nil
	}
	return featureFlagsMap
}

func GetAuthTokenFromHeader(r *http.Request) (tokenStr string, err error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("authorization header is required")
	}
	tokenStr = auth[len("Bearer "):]
	if tokenStr == "" {
		return "", errors.New("invalid token failed to split from bearer")
	}
	return tokenStr, nil
}

func Retry(ctx context.Context, maxRetries int, fn func(ctx context.Context) error) error {
	numbOfRetry := 0

	for numbOfRetry < maxRetries {
		err := fn(ctx)
		if err == nil {
			return nil
		}

		numbOfRetry++
		log.Printf("Retry %d/%d failed: %v", numbOfRetry, maxRetries, err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Second):
		}
	}

	log.Println("⚠️ Max retries reached")
	return fmt.Errorf("max retries reached")
}

func MutexLock(ctx context.Context, fn func(ctx context.Context) error) error {
	mutex := sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	if err := fn(ctx); err != nil {
		return err
	}

	return nil
}

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	letters := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateTransactionID(identifier string) string {
	timeNow := time.Now()
	timeStr := timeNow.Format("20060102150405")
	timeStr = timeStr[:8]
	randomStr := RandomString(6)
	return fmt.Sprintf("%s%s%s", timeStr, identifier, randomStr)
}

func GenerateSubscriptionID() string {
	timeNow := time.Now()
	timeStr := timeNow.Format("20060102150405")
	randomStr := RandomString(12)
	return fmt.Sprintf("%s%s", timeStr, randomStr)
}
