package builder

import (
	"fmt"

	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// List of paginated thread previews for a board
func QrStrLookupThreads(boardID primitive.ObjectID, cfg *types.QueryCtx) (bson.A, error) {
	if boardID == primitive.NilObjectID {
		fmt.Println("Cannot lookup threads with invalid board")
		return nil, fmt.Errorf("invalid board")
	}

	innerMatch := bson.D{{Key: "board", Value: boardID}}
	innerMatch = append(innerMatch, cfg.Search...)

	sortField := cfg.Sort
	if sortField == "" {
		sortField = "updated_at"
	}

	return bson.A{
		BsonD("$match", innerMatch),
		BsonOperator("$sort", sortField, cfg.Order),
		BsonD("$skip", cfg.Skip),
		BsonD("$limit", cfg.Limit),
		BsonOperator("$addFields", "post_count", BsonD("$size", "$posts")),
		QrStrLookupPosts("post_number", -1, 5),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		QrStrLookupIdentity("mods"),
		BsonOperWithArray("$unset", []interface{}{"board", "account", "creator._id", "mods._id"}),
		QrStrLookupAssets(),
	}, nil
}

// a single thread with all posts populated
func QrStrEntireThread(slug string, cfg *types.QueryCtx) bson.A {
	return bson.A{
		BsonOperator("$match", "slug", slug),
		QrStrLookupPosts("post_number", 1, 0),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		QrStrLookupIdentity("mods"),
		BsonOperWithArray("$unset", []interface{}{"board", "account"}),
		QrStrLookupAssets(),
	}
}
