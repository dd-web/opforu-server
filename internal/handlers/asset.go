package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/dd-web/opforu-server/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if rc.UnresolvedAccount {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	/*
	 * Setup variable necessary for all steps, check that the file is valid
	 * and setup temp file for upload
	 */

	file, fileHeader, err := rc.Request.FormFile("file")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}
	defer file.Close()
	details := types.ParseFormFileDetails(rc.Request)

	tmp, err := utils.NewTempAsset(file, fileHeader, details.AssetType.String())
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	fsize, _ := types.GetFileSize(tmp.Dir)
	details.FileSize = uint32(fsize)

	switch details.AssetType {
	case types.AssetTypeImage:
		if rc.Request.ContentLength > types.MAX_FILE_SIZE_IMAGE {
			return ResolveResponseErr(rc, types.ErrorInvalid("file too large"))
		}
	case types.AssetTypeVideo:
		if rc.Request.ContentLength > types.MAX_FILE_SIZE_VIDEO {
			return ResolveResponseErr(rc, types.ErrorInvalid("file too large"))
		}
	}

	/*
	 * Collision detection
	 * checks md5 and sha256 checksums of the file to see if it already exists in the database
	 * we don't need to upload it again if it already exists, and the CDN can serve it.
	 */

	checksummd5, err := types.GetFileChecksumMD5(tmp.Dir)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	checksumsha256, err := types.GetFileChecksumSHA256(tmp.Dir)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	collisionQrStrMd5 := builder.QrStrFindHashCollision(checksummd5, "md5")
	collisionQrStrSha256 := builder.QrStrFindHashCollision(checksumsha256, "sha256")

	collided, err := ah.rh.Store.AssetHashCollisionResolver(collisionQrStrMd5, collisionQrStrSha256)
	if err != nil {
		fmt.Println("error resolving hash collision, probably not a problem tho", err)
	}

	/*
	 * If a collision is detected then we can send back the id of the existing asset
	 * if not then we should make a new asset source here and send the id of that back
	 *
	 * Asset's aren't made until after the user fully submits the new thread/post/whatever
	 * so we don't need to worry about that here
	 */

	if collided != nil {
		rc.AddToResponseList("source_id", collided.ID)
		rc.AddToResponseList("local_id", details.LocalID)

		// add the uploader to list of uploaders if they aren't already in it
		uploaderExistsInList := false
		for _, v := range collided.Uploaders {
			if v == rc.AccountCtx.Account.ID {
				uploaderExistsInList = true
				break
			}
		}

		if !uploaderExistsInList {
			collided.Uploaders = append(collided.Uploaders, rc.AccountCtx.Account.ID)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			collection := ah.rh.Store.DB.Collection("asset_sources")
			opts := options.Update().SetUpsert(false)

			updateQry := builder.BsonOperator("$set", "uploaders", collided.Uploaders)

			_, err := collection.UpdateByID(ctx, collided.ID, updateQry, opts)
			if err != nil {
				fmt.Println("error updating asset source", err)
			}
		}

	} else {
		result, err := types.UploadFileToSpaces(tmp)
		if err != nil {
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}

		assetSrc := types.NewSourceAsset()
		assetSrc.Uploaders = append(assetSrc.Uploaders, rc.AccountCtx.Account.ID)
		assetSrc.AssetType = details.AssetType

		assetSrc.Details.Source.ServerFileName = fmt.Sprintf("%d", tmp.TimeStamp)
		assetSrc.Details.Source.Height = uint16(details.Height)
		assetSrc.Details.Source.Width = uint16(details.Width)
		assetSrc.Details.Source.FileSize = uint32(details.FileSize)
		assetSrc.Details.Source.URL = result
		assetSrc.Details.Source.Extension = tmp.Ext
		assetSrc.Details.Source.HashMD5 = checksummd5
		assetSrc.Details.Source.HashSHA256 = checksumsha256

		// for the time being we're just going to use source files for everything
		// and implement something to create avatars later.
		// we should pretend as if we already have them though. to make updating easier.
		assetSrc.Details.Avatar.ServerFileName = fmt.Sprintf("a-%d", tmp.TimeStamp)

		err = rc.Store.SaveNewSingle(assetSrc, "asset_sources")
		if err != nil {
			fmt.Println("error saving new asset source", err)
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}

		rc.AddToResponseList("source_id", assetSrc.ID)
		rc.AddToResponseList("local_id", details.LocalID)
	}

	// cleanup
	_ = os.Remove(tmp.Dir)

	return ResolveResponse(rc)
}
