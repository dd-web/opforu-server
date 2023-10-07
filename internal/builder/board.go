package builder

import "go.mongodb.org/mongo-driver/bson"

// returns an bson array of bson.D key/value pairs that together construct an aggregation query for
// mongo driver aggregation pipeline
// takes in the board's short name, page number and count per page
func QrStrPublicBoard(bst string, pageNum, pageSize int) bson.A {
	skip := (pageNum - 1) * pageSize
	return bson.A{
		BsonOperator("$match", "short", bst),
		QrStrLookupThread("updated_at", -1, skip, pageSize),
	}
}
