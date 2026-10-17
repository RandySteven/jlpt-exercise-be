package utils

import (
	"fmt"
	"os"
	"reflect"
	"strings"
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
