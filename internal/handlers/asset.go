package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

type AssetHandler struct {
	rh *types.RoutingHandler
}

func InitAssetHandler(rh *types.RoutingHandler) *AssetHandler {
	return &AssetHandler{
		rh: rh,
	}
}

/***********************************************************************************************/
/* ROOT path: host.com/api/assets
/***********************************************************************************************/
func (ah *AssetHandler) RegisterAssetRoot(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return ah.handleAssetList(rc)
	case "POST":
		return ah.handleNewAsset(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/assets
func (ah *AssetHandler) handleAssetList(rc *types.RequestCtx) error {
	return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"message": "asset handler"}, rc)
}

// METHOD: POST
// PATH: host.com/api/assets
func (ah *AssetHandler) handleNewAsset(rc *types.RequestCtx) error {
	fmt.Println("REQUEST:", rc.Request.Body)
	fmt.Println("ContentLength", rc.Request.ContentLength)

	details := types.ParseFormFileDetails(rc.Request)
	fmt.Println("details", details)

	switch details.AssetType {
	case types.AssetTypeImage:
		fmt.Println("image")
		if rc.Request.ContentLength > types.MAX_FILE_SIZE_IMAGE {
			return ResolveResponseErr(rc, types.ErrorInvalid("file too large"))
		}

		file, fileHeader, err := rc.Request.FormFile("file")
		if err != nil {
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}
		defer file.Close()

		err = os.MkdirAll("./tmp/images", os.ModePerm)
		if err != nil {
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}

		dst, err := os.Create(fmt.Sprintf("./tmp/images/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		if err != nil {
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}

	case types.AssetTypeVideo:
		fmt.Println("video")
		if rc.Request.ContentLength > types.MAX_FILE_SIZE_VIDEO {
			return ResolveResponseErr(rc, types.ErrorInvalid("file too large"))
		}
	}

	return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"message": "asset handler"}, rc)
}
