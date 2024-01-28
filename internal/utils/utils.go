package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

// constructs a connection string from env vars
func ParseURIFromEnv() string {
	if os.Getenv("MONGO_URI") != "" {
		return os.Getenv("MONGO_URI")
	}

	return fmt.Sprintf("mongodb://%s:%s/",
		AssertEnvStr(os.Getenv("DB_HOST")),
		AssertEnvStr(os.Getenv("DB_PORT")),
	)
}

// ensures the required string has a value
func AssertEnvStr(v string) string {
	if v == "" {
		log.Fatal("Invalid Environemtn Variable")
	}
	return v
}

// create and return a new aws config
func NewS3Config() *aws.Config {
	key := os.Getenv("DO_API_KEY")
	secret := os.Getenv("DO_API_SECRET")
	// endpoint := os.Getenv("DO_API_ENDPOINT")
	// region := os.Getenv("DO_API_REGION")

	return &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String("nyc3.digitaloceanspaces.com"),
		Region:           aws.String("nyc3"),
		S3ForcePathStyle: aws.Bool(false),
	}
}
