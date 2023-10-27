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
	rc *types.RequestCtx
	rh *types.RoutingHandler
}

func InitAccountHandlers(rh *types.RoutingHandler) *AccountHandler {
	return &AccountHandler{
		rh: rh,
	}
}

func (ah *AccountHandler) UpdateCtx(rc *types.RequestCtx) {
	ah.rc = rc
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountRoot(rc *types.RequestCtx) error {
	ah.UpdateCtx(rc)
	// qCfg := utils.NewQueryConfig(r, "accounts")

	switch rc.Request.Method {
	case "GET":
		return ah.handleGetAccount(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/account
func (ah *AccountHandler) handleGetAccount(rc *types.RequestCtx) error {
	// fmt.Println("Account handler", q)
	return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"message": "account handler"})
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/login
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountLogin(rc *types.RequestCtx) error {
	ah.UpdateCtx(rc)
	// qCfg := utils.NewQueryConfig(r, "accounts")
	fmt.Println("HELLOO!??")

	switch rc.Request.Method {
	case "POST":
		return ah.handlePostAccountLogin(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// POST: host.com/api/account/login
func (ah *AccountHandler) handlePostAccountLogin(rc *types.RequestCtx) error {
	ah.UpdateCtx(rc)

	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	// the username field could hold either the username or email
	var parsed struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println("Error parsing body", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	account, err := rc.Store.FindAccountByUsernameOrEmail(parsed.Username, "")
	if err != nil {
		fmt.Println("Error finding account (find by username)", err)
		return HandleSendJSON(rc.Writer, http.StatusUnauthorized, bson.M{"error": "invalid login credentials"})
	}

	passwordMatches := utils.CompareHash(account.Password, parsed.Password)
	if !passwordMatches {
		fmt.Println("Password mismatch")
		return HandleSendJSON(rc.Writer, http.StatusUnauthorized, bson.M{"error": "invalid login credentials"})
	}

	session := types.NewSession(account.ID)

	err = ah.rh.Store.SaveNewSingle(session, "sessions")
	if err != nil {
		fmt.Println("Error saving new session", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"})
	}

	resp := bson.M{
		"account": account.FormatForClient(),
		"session": session.FormatForClient(),
	}

	return HandleSendJSON(rc.Writer, http.StatusOK, resp)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/register
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountRegister(rc *types.RequestCtx) error {
	ah.UpdateCtx(rc)
	// qCfg := utils.NewQueryConfig(r, "accounts")

	switch rc.Request.Method {
	case "POST":
		return ah.handlePostAccountRegister(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// POST: host.com/api/account/register
func (ah *AccountHandler) handlePostAccountRegister(rc *types.RequestCtx) error {
	ah.UpdateCtx(rc)

	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
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
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	account, err := ah.rh.Store.FindAccountByUsernameOrEmail(parsed.Username, parsed.Email)
	if err != nil {
		fmt.Println("This isn't really an error, this is what we want (no accounts with existing username or email)", err)
	}

	if account != nil {
		fmt.Println("Account already exists")
		return HandleSendJSON(rc.Writer, http.StatusBadRequest, bson.M{"error": "account already exists"})
	}

	newAccount := types.NewAccount()
	newAccount.Username = parsed.Username
	newAccount.Email = parsed.Email

	session := types.NewSession(newAccount.ID)

	err = ah.rh.Store.SaveNewSingle(newAccount, "accounts")
	if err != nil {
		fmt.Println("Error saving new account", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"})
	}

	err = ah.rh.Store.SaveNewSingle(session, "sessions")
	if err != nil {
		fmt.Println("Error saving new session", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"})
	}

	resp := bson.M{
		"account": newAccount.FormatForClient(),
		"session": session.FormatForClient(),
	}

	return HandleSendJSON(rc.Writer, http.StatusOK, resp)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/me
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountMe(rc *types.RequestCtx) error {
	ah.UpdateCtx(rc)
	// qCfg := utils.NewQueryConfig(, "accounts")

	switch rc.Request.Method {
	case "POST":
		return ah.handlePostAccountMe(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// POST: host.com/api/account/me
func (ah *AccountHandler) handlePostAccountMe(rc *types.RequestCtx) error {
	ah.UpdateCtx(rc)
	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		fmt.Println("Error reading body", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	var parsed struct {
		Session string `json:"session_id"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println("Error parsing body", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid request body"})
	}

	session, err := rc.Store.FindSession(parsed.Session)
	if err != nil {
		// fmt.Println("Error finding session", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid session"})
	}

	if session.IsExpired() {
		fmt.Println("Session expired")
		return HandleSendJSON(rc.Writer, http.StatusUnauthorized, bson.M{"error": "session expired"})
	}

	account, err := rc.Store.FindAccountByID(session.AccountID)
	if err != nil {
		fmt.Println("Error finding account", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "invalid session"})
	}

	resp := bson.M{
		"account": account.FormatForClient(),
		"session": session.FormatForClient(),
	}

	return HandleSendJSON(rc.Writer, http.StatusOK, resp)

}
