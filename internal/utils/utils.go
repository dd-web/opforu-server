package utils

import (
	"crypto/aes"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
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

// EncryptAES encrypts a string using AES encryption
func EncryptAES(plaintext string, password string) (string, error) {
	cypher, err := aes.NewCipher([]byte(password))
	if err != nil {
		return "", err
	}

	bs := make([]byte, len(plaintext))
	cypher.Encrypt(bs, []byte(plaintext))

	return hex.EncodeToString(bs), nil
}

// DecryptAES decrypts a string using AES encryption
func DecryptAES(encrypted string, password string) (string, error) {
	decoded, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	cypher, err := aes.NewCipher([]byte(password))
	if err != nil {
		return "", err
	}

	bs := make([]byte, len(decoded))
	cypher.Decrypt(bs, decoded)

	return string(bs), nil
}

// Returns the character set used for generating identity IDs
func GetIdentityCharSet() string {
	return "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789-_"
}

// Returns the character set used for generating thread slugs
func GetThreadSlugCharSet() string {
	return "abcdefghijklmnopqrstuvwxyz0123456789-_"
}

type QueryConfig struct {
	Sort     bson.M
	Limit    int
	Skip     int
	Filter   bson.M
	PageInfo *PageConfig
}

// constructs a new query config object from the request including page details
func NewQueryConfig(r *http.Request) (*QueryConfig, error) {
	p, err := NewPageConfig(r)
	if err != nil {
		return nil, err
	}

	return &QueryConfig{
		Sort:     bson.M{},
		Limit:    p.PageSize,
		Skip:     0,
		Filter:   bson.M{},
		PageInfo: p,
	}, nil
}

// helps paginate data
type PageConfig struct {
	Current  int `json:"current_page"` // current page number
	PageSize int `json:"page_size"`    // number of records per page
	Total    int `json:"total_pages"`  // total number of pages

	Records      []any `json:"records"`       // data the page contains
	TotalRecords int   `json:"total_records"` // total number of records (determines total number of pages)

	IsLast        bool `json:"last_page"`      // is this the last page
	LastPageCount int  `json:"last_page_size"` // number of records on the last page
}

// analyzes the request and constructs a new page config object from it
func NewPageConfig(r *http.Request) (*PageConfig, error) {
	q := r.URL.Query()

	current, err := strconv.Atoi(q.Get("page"))
	if err != nil {
		current = 1
	}

	size, err := strconv.Atoi(q.Get("count"))
	if err != nil {
		size = 10
	}

	return &PageConfig{
		Current:  current,
		PageSize: size,
	}, nil
}

// takes in the total number of records (in the database) and updates the page config to reflect that
func (p *PageConfig) Update(count int) {
	pmod := count % p.PageSize

	p.TotalRecords = count
	p.Total = count / p.PageSize
	p.LastPageCount = pmod

	if pmod > 0 {
		p.Total++
	}

	if p.Current == p.Total || p.Current >= p.Total || p.Total == 0 {
		p.IsLast = true
	}
}
