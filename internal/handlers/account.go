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

type AccountHandler struct {
	rh *types.RoutingHandler
}

func InitAccountHandlers(rh *types.RoutingHandler) *AccountHandler {
	return &AccountHandler{
		rh: rh,
	}
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountRoot(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return ah.handleGetAccount(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/account
func (ah *AccountHandler) handleGetAccount(rc *types.RequestCtx) error {
	return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"message": "account handler"}, rc)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/login
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountLogin(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "POST":
		return ah.handlePostAccountLogin(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: POST
// PATH: host.com/api/account/login
func (ah *AccountHandler) handlePostAccountLogin(rc *types.RequestCtx) error {
	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"}, rc)
	}

	// the username field could hold either the username or email
	var parsed struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println("Error parsing body", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"}, rc)
	}

	account, err := rc.Store.FindAccountByUsernameOrEmail(parsed.Username, "")
	if err != nil {
		fmt.Println("Error finding account (find by username)", err)
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
		// return HandleSendJSON(rc.Writer, http.StatusUnauthorized, bson.M{"error": "invalid login credentials"}, rc)
	}

	passwordMatches := utils.CompareHash(account.Password, parsed.Password)
	if !passwordMatches {
		fmt.Println("Password mismatch")
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
		// return HandleSendJSON(rc.Writer, http.StatusUnauthorized, bson.M{"error": "invalid login credentials"}, rc)
	}

	session := types.NewSession(account)

	err = rc.Store.SaveNewSingle(session, "sessions")
	if err != nil {
		fmt.Println("Error saving new session", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"}, rc)
	}

	rc.AccountCtx.Account = account
	rc.AccountCtx.Session = session

	return ResolveResponse(rc)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/register
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountRegister(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "POST":
		return ah.handlePostAccountRegister(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: POST
// PATH: host.com/api/account/register
func (ah *AccountHandler) handlePostAccountRegister(rc *types.RequestCtx) error {
	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"}, rc)
	}

	var parsed struct {
		Username        string `json:"username"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println("Error parsing body", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"}, rc)
	}

	account, _ := rc.Store.FindAccountByUsernameOrEmail(parsed.Username, parsed.Email)
	if account != nil {
		fmt.Println("Account already exists")
		return ResolveResponseErr(rc, types.ErrorConflict("email or username already exists"))
		// return HandleSendJSON(rc.Writer, http.StatusBadRequest, bson.M{"error": "account already exists"}, rc)
	}

	pwh, err := utils.HashPassword(parsed.Password)
	if err != nil {
		fmt.Println("Error hashing password", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"}, rc)
	}

	newAccount := types.NewAccount()
	newAccount.Username = parsed.Username
	newAccount.Email = parsed.Email
	newAccount.Password = pwh

	session := types.NewSession(newAccount)

	err = rc.Store.SaveNewSingle(newAccount, "accounts")
	if err != nil {
		fmt.Println("Error saving new account", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"}, rc)
	}

	err = rc.Store.SaveNewSingle(session, "sessions")
	if err != nil {
		fmt.Println("Error saving new session", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
		// return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"}, rc)
	}

	rc.AccountCtx.Account = newAccount
	rc.AccountCtx.Session = session

	return ResolveResponse(rc)
}
