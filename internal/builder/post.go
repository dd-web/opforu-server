package builder

import "go.mongodb.org/mongo-driver/bson"

// returns the bson.D pipeline obj to unflatten the creator field back into an object
func QrStrAddCreatorOnThread() bson.D {
	arrElmAt := BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})
	return BsonOperator("$addFields", "creator", arrElmAt)
}

// Post lookup aggregation pipeline
func QrStrLookupPosts(sortBy string, sortDir int, limit int) bson.D {
	pipe := bson.A{
		BsonOperator("$sort", sortBy, sortDir),
		BsonD("$limit", limit),
		// lookup media here
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		BsonOperWithArray("$unset", []interface{}{"thread", "board", "account"}),
	}

	return BsonLookup("posts", "posts", "_id", "posts", bson.D{}, pipe)
}
