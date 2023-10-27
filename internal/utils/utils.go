package utils

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var envKeys = []string{
	"env", "ENV", "environment", "ENVIRONMENT", "node_env", "NODE_ENV",
}

var envVals = []string{
	"prod", "production", "PROD", "PRODUCTION",
}

// cross reference env keys and values to determine if we are in production
func IsProdEnv() bool {
	for _, k := range envKeys {
		val := os.Getenv(k)
		if val == "" {
			continue
		}
		for _, v := range envVals {
			if val == v {
				return true
			}
		}
	}
	return false
}

// ensures the passed string has a value and is not empty
func AssertString(v string) string {
	if v == "" {
		panic("mandatory string value is empty")
	}
	return v
}

// hashes a password with bcrypt
func HashPassword(plaintext string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// compare a hashed password with plaintext
func CompareHash(hashed, plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
