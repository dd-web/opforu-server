package utils

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

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
	endpoint := os.Getenv("DO_API_ENDPOINT")
	region := os.Getenv("DO_API_REGION")

	return &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(false),
	}
}

type TempAsset struct {
	TimeStamp int64
	AssetType string
	Ext       string
	Dir       string
}

func NewTempAsset(file multipart.File, header *multipart.FileHeader, kind string) (*TempAsset, error) {
	tempdir := "./tmp/" + kind + "s/"

	err := os.MkdirAll(tempdir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixNano()
	ext := filepath.Ext(header.Filename)
	tempf := fmt.Sprintf(tempdir+"%d%s", now, ext)

	tempFile, err := os.Create(tempf)
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		return nil, err
	}

	return &TempAsset{
		TimeStamp: now,
		AssetType: kind,
		Ext:       ext,
		Dir:       tempf,
	}, nil

}
