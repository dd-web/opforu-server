package builder

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// returns the bson.D pipeline obj to unflatten the creator field back into an object
func QrStrAddCreatorOnThread() bson.D {
	arrElmAt := BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})
	return BsonOperator("$addFields", "creator", arrElmAt)
}

// Thread posts aggregation lookup pipeline
func QrStrLookupPosts(sortBy string, sortDir int, limit int) bson.D {
	pipe := bson.A{
		BsonOperator("$sort", sortBy, sortDir),
	}

	if limit > 0 {
		pipe = append(pipe, BsonD("$limit", limit))
	}

	pipe = append(
		pipe,
		QrStrLookupAssets("assets"),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		BsonOperWithArray("$unset", []interface{}{"thread", "board", "account", "creator._id"}),
	)

	return BsonLookup("posts", "posts", "_id", "posts", bson.D{}, pipe)
}

// singular post lookup
func QrStrLookupPost(threadID primitive.ObjectID, postNum int, threadSlug, boardShort string) bson.A {
	pipe := bson.D{}
	pipe = append(pipe, BsonE("thread", threadID))
	pipe = append(pipe, BsonE("post_number", postNum))

	return bson.A{
		BsonD("$match", pipe),
		BsonD("$limit", 1),
		QrStrLookupAssets("assets"),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		BsonOperator("$addFields", "thread", threadSlug),
		BsonOperator("$addFields", "board", boardShort),
		BsonOperWithArray("$unset", []interface{}{"account", "_id", "creator._id"}),
	}
}
