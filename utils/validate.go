package utils

import (
	"fmt"
	"github.com/RandySteven/go-kopi/queries"
	"golang.org/x/crypto/bcrypt"
	"path"
	"strings"
)

func QueryValidation(query queries.GoQuery, command string) error {
	queryStr := query.ToString()
	if !strings.Contains(queryStr, command) {
		return fmt.Errorf(`the query command is not valid`)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hash), err
}

func ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateFileType(fileName string) bool {
	extension := path.Ext(fileName)
	imageFileExts := []string{
		"jpg",
		"jpeg",
		"png",
	}

	flag := false

	for _, imageFileExt := range imageFileExts {
		if extension == imageFileExt {
			flag = true
		}
	}

	if flag == true {
		return true
	}

	return false
}
