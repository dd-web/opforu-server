package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dd-web/opforu-server/internal/types"
	"github.com/dd-web/opforu-server/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func AccountHandler(w http.ResponseWriter, r *http.Request) error {
	return nil
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterAccountRoot(w http.ResponseWriter, r *http.Request) error {
	qCfg := utils.NewQueryConfig(r, "accounts")

	switch r.Method {
	case "GET":
		return rh.handleGetAccount(w, r, qCfg)
	default:
		return HandleUnsupportedMethod(w, r)
	}
}

// GET: host.com/api/account
func (rh *RoutingHandler) handleGetAccount(w http.ResponseWriter, r *http.Request, q *utils.QueryConfig) error {
	fmt.Println("Account handler", q)
	return HandleSendJSON(w, http.StatusOK, bson.M{"message": "account handler"})
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/login
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterAccountLogin(w http.ResponseWriter, r *http.Request) error {
	qCfg := utils.NewQueryConfig(r, "accounts")

	switch r.Method {
	case "GET":
		return rh.handleGetAccountLogin(w, r, qCfg)
	case "POST":
		return rh.handlePostAccountLogin(w, r, qCfg)
	default:
		return HandleUnsupportedMethod(w, r)
	}
}

// GET: host.com/api/account/login
func (rh *RoutingHandler) handleGetAccountLogin(w http.ResponseWriter, r *http.Request, q *utils.QueryConfig) error {
	fmt.Println("Account login handler GET", q)
	return HandleSendJSON(w, http.StatusOK, bson.M{"message": "account handler"})
}

// POST: host.com/api/account/login
func (rh *RoutingHandler) handlePostAccountLogin(w http.ResponseWriter, r *http.Request, q *utils.QueryConfig) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return HandleSendJSON(w, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	var parsed struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println("Error parsing body", err)
		return HandleSendJSON(w, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	account, err := rh.Store.FindAccountByUsernameOrEmail(parsed.Username)
	if err != nil {
		fmt.Println("Error finding account (find by username)", err)
		return HandleSendJSON(w, http.StatusUnauthorized, bson.M{"error": "invalid login credentials"})
	}

	passwordMatches := utils.CompareHash(account.Password, parsed.Password)
	if !passwordMatches {
		fmt.Println("Password mismatch")
		return HandleSendJSON(w, http.StatusUnauthorized, bson.M{"error": "invalid login credentials"})
	}

	session := types.NewSession(account.ID)

	err = rh.Store.SaveNewSingle(session, "sessions")
	if err != nil {
		fmt.Println("Error saving new session", err)
		return HandleSendJSON(w, http.StatusInternalServerError, bson.M{"error": "unexpected server error"})
	}

	resp := bson.M{
		"account": account.FormatForClient(),
		"session": session.FormatForClient(),
	}

	return HandleSendJSON(w, http.StatusOK, resp)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/me
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterAccountMe(w http.ResponseWriter, r *http.Request) error {
	qCfg := utils.NewQueryConfig(r, "accounts")

	switch r.Method {
	case "POST":
		return rh.handlePostAccountMe(w, r, qCfg)
	default:
		return HandleUnsupportedMethod(w, r)
	}
}

// POST: host.com/api/account/me
func (rh *RoutingHandler) handlePostAccountMe(w http.ResponseWriter, r *http.Request, q *utils.QueryConfig) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return HandleSendJSON(w, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	var parsed struct {
		Session string `json:"session"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println("Error parsing body", err)
		return HandleSendJSON(w, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	session, err := rh.Store.FindSession(parsed.Session)
	if err != nil {
		fmt.Println("Error finding session", err)
		return HandleSendJSON(w, http.StatusInternalServerError, bson.M{"error": "invalid session"})
	}

	if session.IsExpired() {
		fmt.Println("Session expired")
		return HandleSendJSON(w, http.StatusUnauthorized, bson.M{"error": "session expired"})
	}

	account, err := rh.Store.FindAccountByID(session.AccountID)
	if err != nil {
		fmt.Println("Error finding account", err)
		return HandleSendJSON(w, http.StatusInternalServerError, bson.M{"error": "invalid session"})
	}

	resp := bson.M{
		"account": account.FormatForClient(),
		"session": session.FormatForClient(),
	}

	return HandleSendJSON(w, http.StatusOK, resp)

}
