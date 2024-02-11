package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dd-web/opforu-server/internal/types"
	"github.com/dd-web/opforu-server/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	// the username field could hold either the username or email
	var parsed struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	account, err := rc.Store.FindAccountByUsernameOrEmail(parsed.Username, "")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	passwordMatches := utils.CompareHash(account.Password, parsed.Password)
	if !passwordMatches {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	session := types.NewSession(account)

	err = rc.Store.SaveNewSingle(session, "sessions")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
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
	}

	account, _ := rc.Store.FindAccountByUsernameOrEmail(parsed.Username, parsed.Email)
	if account != nil {
		fmt.Println("Account already exists")
		return ResolveResponseErr(rc, types.ErrorConflict("email or username already exists"))
	}

	pwh, err := utils.HashPassword(parsed.Password)
	if err != nil {
		fmt.Println("Error hashing password", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())
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
	}

	err = rc.Store.SaveNewSingle(session, "sessions")
	if err != nil {
		fmt.Println("Error saving new session", err)
		return ResolveResponseErr(rc, types.ErrorUnexpected())

	}

	rc.AccountCtx.Account = newAccount
	rc.AccountCtx.Session = session

	return ResolveResponse(rc)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/logout
/***********************************************************************************************/
func (ah *AccountHandler) RegisterAccountLogout(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "POST":
		return ah.handlePostAccountLogout(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: POST
// PATH: host.com/api/account/logout
func (ah *AccountHandler) handlePostAccountLogout(rc *types.RequestCtx) error {
	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	var parsed struct {
		SessionID string `json:"session_id"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	session, err := rc.Store.FindSession(parsed.SessionID)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	err = rc.Store.DeleteSingle(session.ID, "sessions")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	for k := range rc.Store.Cache.Sessions {
		if k == parsed.SessionID {
			delete(rc.Store.Cache.Sessions, k)
		}
	}

	// bypass the response resolver so it doesn't auto populate the deleted session
	return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"message": "logged out"}, rc)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/account/favorites
/***********************************************************************************************/

func (ah *AccountHandler) RegisterAccountFavoriteRoot(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return ah.handleGetFavoriteList(rc)
	case "POST":
		return ah.handleAddRemoveFavorite(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/account/favorites
func (ah *AccountHandler) handleGetFavoriteList(rc *types.RequestCtx) error {
	if rc.UnresolvedAccount {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	list, err := rc.Store.FindAccountFavoriteAssetList(rc.AccountCtx.Account.ID)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("list not found"))
	}

	// @TODO: implement asset aggregation pipeline from here. as of now we'll just get a list of objectID's
	rc.AddToResponseList("favorite_assets", list.Items)

	return ResolveResponse(rc)
}

// METHOD: POST
// PATH: host.com/api/account/favorites
func (ah *AccountHandler) handleAddRemoveFavorite(rc *types.RequestCtx) error {
	if rc.UnresolvedAccount {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	ts := time.Now().UTC()
	RRMOper := types.NewRRMFavoriteAsset()
	details := types.NewRUMFavoriteAsset()

	// get the asset list of the account
	list, err := rc.Store.FindAccountFavoriteAssetList(rc.AccountCtx.Account.ID)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("list not found"))
	}

	// find the asset that we want to add/remove
	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = json.Unmarshal(body, &details)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	asset, err := rc.Store.FindAssetByID(details.AssetID)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	// if id is in list - remove it.
	RRMOper.AssetID = asset.ID
	itemExistsInList := false
	newItemList := []primitive.ObjectID{}

	for _, v := range list.Items {
		if v == asset.ID {
			itemExistsInList = true
		} else {
			newItemList = append(newItemList, v)
		}
	}

	if itemExistsInList {
		RRMOper.Oper = types.RRMFAOperRemove
	} else {
		RRMOper.Oper = types.RRMFAOperAdd
		// if id not in list -
		//	if account_id on asset is the called - add it to the list
		if asset.AccountID == rc.AccountCtx.Account.ID {
			newItemList = append(newItemList, asset.ID)
		} else {
			// if account_id is different, create a new asset for the current account and add it to the list.
			newAsset := types.CloneAsset(asset, rc.AccountCtx.Account.ID)
			RRMOper.AssetID = newAsset.ID
			newAsset.UpdatedAt = &ts
			newItemList = append(newItemList, newAsset.ID)

			err = rc.Store.SaveNewSingle(newAsset, "asset")
			if err != nil {
				return ResolveResponseErr(rc, types.ErrorUnexpected())
			}
		}
	}

	list.Items = newItemList
	list.UpdatedAt = &ts
	err = rc.Store.UpdateFavoriteAssetList(list)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	rc.AddToResponseList("details", RRMOper)
	return ResolveResponse(rc)
}
