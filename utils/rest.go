package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/iancoleman/strcase"
)

func ContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func BindJSON(req *http.Request, request interface{}) error {
	return json.NewDecoder(req.Body).Decode(&request)
}

func BindForm(req *http.Request, request any) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}
	bindForm := make(map[string]interface{})
	for key, values := range req.Form {
		log.Printf("key : %s value : %s", key, values[0])
		bindForm[key] = values[0]
	}

	fields := GetFieldsOfObject(request)

	for _, field := range fields {
		val := reflect.ValueOf(bindForm[strcase.ToSnake(field)])
		if val.IsValid() {
			fieldValue := reflect.ValueOf(request).Elem().FieldByName(field)
			if fieldValue.IsValid() && fieldValue.CanSet() {
				if val.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(val)
				} else {
					return fmt.Errorf("type conversion error for field %s", field)
				}
			}
		}
	}

	return nil
}

func ResponseHandler(w http.ResponseWriter, responseCode int, message string, dataKey *string, responseData any, err error) {
	w.WriteHeader(responseCode)
	responseMap := make(map[string]any)
	if dataKey != nil && responseData != nil {
		responseMap[*dataKey] = responseData
	}
	var response *responses.Response
	if len(responseMap) == 0 {
		response = responses.NewResponseWithoutData(message, err)
	} else {
		response = responses.NewResponse(message, responseMap, err)
	}
	log.Print(response)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		log.Fatal(err)
		return
	}
}
