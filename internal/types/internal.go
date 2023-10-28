package types

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// These define the valid resource paths for the API directly after root.
type APIResource string

// APIResource constants will by default be it's singular form. These will be modified when necessary
// to their plural forms to match the database collection names.
const (
	Resource_Account  APIResource = "account"
	Resource_Article  APIResource = "article"
	Resource_Board    APIResource = "board"
	Resource_Identity APIResource = "identity"
	Resource_Post     APIResource = "post"
	Resource_Thread   APIResource = "thread"
	Resource_Session  APIResource = "session"
	Resource_Asset    APIResource = "asset"
)

// holds all of the resolved/parsed request details and info so that handlers can be more simple and focused.
type RequestCtx struct {
	Request      *http.Request       // request
	Writer       http.ResponseWriter // writer
	Query        *QueryCtx           // the parsed query context (or nil if irrelevant/not yet parsed)
	Resource     APIResource         // this is the main subroute of the API, the first major path after root.
	Pagination   *PageCtx            `json:"pages"`   // pagination information
	Records      []bson.M            `json:"records"` // resource(s) we intend to return to the client
	Store        *Store
	AccountCtx   *AccountCtx
	ResponseList []bson.M
	ResponseData bson.M
	// ResponseFormatted bson.D
}

// creates a new request context - parses and resolves request details into the context
// in order for handlers to be more simple and focused
func NewRequestCtx(w http.ResponseWriter, r *http.Request) *RequestCtx {
	return (&RequestCtx{
		Request:      r,
		Writer:       w,
		Query:        NewQueryCtx(),
		Resource:     APIResource(strings.Split(r.URL.Path, "/")[2]),
		Pagination:   NewPageCtx(),
		Records:      []bson.M{},
		ResponseList: []bson.M{},
		ResponseData: bson.M{},
		AccountCtx:   NewAccountCtx(),
	}).Resolve()
}

// updates the request context with the store
func (rc *RequestCtx) UpdateStore(s *Store) {
	rc.Store = s
}

// adds the given key/value pair to the response list to be returned to the
// client when the response is finalized
func (rc *RequestCtx) AddToResponseList(k string, v any) {
	rc.ResponseList = append(rc.ResponseList, bson.M{k: v})
}

// prepares the response list to be sent to the client
func (rc *RequestCtx) Finalize() {
	rc.AddToResponseList("paginator", rc.Pagination)
	rc.AddToResponseList("records", rc.Records)

	for _, v := range rc.ResponseList {
		for key, value := range v {
			rc.ResponseData[key] = value
		}
	}
}

// parse the request and populate each of the contexts with relevant information
// certain contexts must be done synchronously in a certain order to ensure the
// necessary data is available for the next context to be resolved
func (rc *RequestCtx) Resolve() *RequestCtx {
	var current_page int = 1
	var page_size int = 10

	if rc.Request != nil {
		RequestQuery := rc.Request.URL.Query()

		for k, v := range RequestQuery {
			switch k {

			case "page":
				current, err := strconv.Atoi(v[0])
				if err != nil {
					break
				}
				current_page = current

			case "count":
				size, err := strconv.Atoi(v[0])
				if err != nil {
					break
				}
				page_size = size

			case "order":
				order, err := strconv.Atoi(v[0])
				if err != nil {
					break
				}
				rc.Query.Order = order

			case "sort":
				rc.Query.Sort = v[0]

			case "search":
				rc.Query.Search = bson.D{{
					Key: rc.Query.Sort, Value: bson.D{{
						Key: "$regex", Value: primitive.Regex{Pattern: v[0], Options: "i"},
					}},
				}}

			default:
				rc.Query.UnhandledQueryParams[k] = v[0] // unknown query param
			}

		}

		rc.Query.Skip = int64((current_page - 1) * page_size)
		rc.Query.Limit = int64(page_size)

	}

	var filter bson.D = bson.D{}
	for k, v := range rc.Query.UnhandledQueryParams {
		filter = append(filter, bson.E{Key: k, Value: v})
	}

	rc.Query.Filter = filter
	rc.Pagination.Current = current_page
	rc.Pagination.Count = page_size

	return rc
}

// request query context information
type QueryCtx struct {
	Sort                 string         // field to sort by
	Order                int            // 1 for ascending, -1 for descending
	Limit                int64          // number of records to return (size of page)
	Skip                 int64          // number of records to skip (page number * page size)
	Search               bson.D         // if we're searching for something
	Filter               bson.D         // if we're filtering for something
	UnhandledQueryParams map[string]any // any query params that we don't know what to do with
}

// creates a new query context with default values
func NewQueryCtx() *QueryCtx {
	return &QueryCtx{
		Sort:                 "title",
		Order:                -1,
		Limit:                10,
		Skip:                 0,
		Search:               bson.D{},
		Filter:               bson.D{},
		UnhandledQueryParams: make(map[string]any),
	}
}

// interim struct to hold pagination information
type PageCtx struct {
	Current   int  `json:"current_page"`   // current page number
	Count     int  `json:"page_size"`      // number of records per page
	Pages     int  `json:"total_pages"`    // total number of pages
	Records   int  `json:"total_records"`  // total number of records (determines total number of pages)
	Last      bool `json:"last_page"`      // is this the last page
	Remainder int  `json:"last_page_size"` // number of records on the last page
}

// creates a new page context with default values
func NewPageCtx() *PageCtx {
	return &PageCtx{
		Current: 1,
		Count:   10,
	}
}

// updates pagination info based on the query and it's results
// NOTE: this needs to be called after query resolver has been called otherwise the object will be empty
func (p *PageCtx) Update(results int) {
	remainder := results % p.Count

	p.Records = results
	p.Pages = results / p.Count
	p.Remainder = remainder

	if remainder > 0 {
		p.Pages++
	}

	if p.Current == p.Pages || p.Current >= p.Pages || p.Pages == 0 {
		p.Last = true
	}
}

// context of the requesting user. if unable to resolve a user pointers will be nil
type AccountCtx struct {
	Session *Session
	Account *Account
	Role    AccountRole
}

// creates a new user context with default values and a public role
func NewAccountCtx() *AccountCtx {
	return &AccountCtx{
		Role: AccountRolePublic,
	}
}

// logs the request to the console
func RequestLogger(rc *RequestCtx) {
	fmt.Printf("[%s]: %s - %s \n", rc.Request.Method, rc.Request.URL.Path, rc.Request.RemoteAddr)
}

// top level router with access to the store for database operations
type RoutingHandler struct {
	Router *mux.Router
	Store  *Store
}

func NewRoutingHandler(s *Store) *RoutingHandler {
	r := mux.NewRouter()
	return &RoutingHandler{
		Router: r,
		Store:  s,
	}
}
