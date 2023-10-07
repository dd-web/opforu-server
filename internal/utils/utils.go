package utils

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Sort                 string
	Order                int
	Limit                int64
	Skip                 int64
	Search               bson.D
	ResourceType         string
	Filter               bson.D
	PageInfo             *PageConfig
	UnhandledQueryParams map[string]any
}

// constructs a new query config object from the request including page details
// also takes a string for the resource type we're querying and paginating
func NewQueryConfig(r *http.Request, rt string) *QueryConfig {
	qc := &QueryConfig{
		Sort:                 "",
		Order:                -1,
		Limit:                10,
		Skip:                 0,
		Search:               bson.D{},
		ResourceType:         rt,
		Filter:               bson.D{},
		UnhandledQueryParams: make(map[string]any),
	}

	var current int = 1
	var size int = 10

	if r != nil {
		URLQuery := r.URL.Query()
		for k, v := range URLQuery {
			switch k {
			case "page":
				currentInt, err := strconv.Atoi(v[0])
				if err != nil {
					fmt.Println("Error converting page to int", err)
					break
				}
				current = currentInt
			case "count":
				sizeInt, err := strconv.Atoi(v[0])
				if err != nil {
					fmt.Println("Error converting count to int", err)
					break
				}
				size = sizeInt
			case "order":
				orderInt, err := strconv.Atoi(v[0])
				if err != nil {
					fmt.Println("Error converting order to int", err)
					break
				}
				qc.Order = orderInt
			case "sort":
				qc.Sort = v[0]
			case "search":
				qc.Search = bson.D{{
					Key: "title", Value: bson.D{{
						Key:   "$regex",
						Value: primitive.Regex{Pattern: v[0]},
					}},
				}}
			default:
				qc.UnhandledQueryParams[k] = v[0]
			}
		}
		qc.Skip = int64((current - 1) * size)
		qc.Limit = int64(size)
	}

	// @TODO: still need to add the constraints from the BSON constructor in the builder
	// because right now it only gets the number of records without those constraints, which
	// will give us invalid page counts -- been at this too long today though, FIX LATER
	var flt bson.D = bson.D{}
	for k, v := range qc.UnhandledQueryParams {
		flt = append(flt, bson.E{Key: k, Value: v})
	}

	qc.Filter = flt

	fmt.Println("Search:", qc.Search)
	// for k, v := range urlFilters {

	// }

	qc.PageInfo = NewPageConfig(r, current, size)

	// fmt.Println("QueryConfig", qc, qc.PageInfo, qc.Filters)
	fmt.Println("Filters:", qc.Filter)

	return qc
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
func NewPageConfig(r *http.Request, current, size int) *PageConfig {
	return &PageConfig{
		Current:  current,
		PageSize: size,
	}
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
