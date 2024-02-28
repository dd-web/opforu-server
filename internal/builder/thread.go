package builder

import (
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// List of paginated thread previews for a board
func QrStrLookupThreads(boardID primitive.ObjectID, cfg *types.QueryCtx) (bson.A, error) {
	pipe, err := StartPaginatedPipe("board", boardID, cfg)
	if err != nil {
		return nil, err
	}

	pipe = append(
		pipe,
		BsonOperator("$addFields", "post_count", BsonD("$size", "$posts")),
		QrStrLookupPosts("post_number", -1, 5),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		QrStrLookupIdentity("mods"),
		BsonOperWithArray("$unset", []interface{}{"board", "account", "creator._id", "mods._id"}),
		QrStrLookupAssets("assets"),
	)

	return pipe, nil
}

// a single thread with all posts populated
func QrStrEntireThread(slug string, cfg *types.QueryCtx) bson.A {
	return bson.A{
		BsonOperator("$match", "slug", slug),
		QrStrLookupPosts("post_number", 1, 0),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		QrStrLookupIdentity("mods"),
		BsonOperWithArray("$unset", []interface{}{"board", "account", "creator._id", "mods._id"}),
		QrStrLookupAssets("assets"),
	}
}

// singular thread lookup - models in the shape of post
func QrStrLookupThread(slug string, boardID primitive.ObjectID, boardShort string) bson.A {
	pipe := bson.D{}
	pipe = append(pipe, BsonE("slug", slug))
	pipe = append(pipe, BsonE("board", boardID))

	return bson.A{
		BsonD("$match", pipe),
		BsonD("$limit", 1),
		QrStrLookupAssets("assets"),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		BsonOperator("$addFields", "thread", slug),
		BsonOperator("$addFields", "board", boardShort),
		BsonOperator("$addFields", "post_count", BsonD("$size", "$posts")),
		BsonOperWithArray("$unset", []interface{}{"account", "_id", "creator._id", "flags", "posts"}),
	}
}
