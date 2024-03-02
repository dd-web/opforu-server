package builder

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	PUBLIC_INCLUDE_FIELDS = []string{"name", "style", "role", "status"}
)

// returns a pipeline object that can be used as an internal pipeline for a $lookup or other aggregate
// pk is the field name to reference the lookup from
func QrStrLookupIdentity(pk string) bson.D {
	projection := bson.A{BsonProjection(PUBLIC_INCLUDE_FIELDS, 1)}
	return BsonLookup("identities", pk, "_id", pk, bson.D{}, projection)
}

// 10 most recent account identities
func QrStrLookupAccountRecentIdentities(account_id primitive.ObjectID) bson.A {
	pipe := bson.D{}
	pipe = append(pipe, BsonE("account", account_id))

	return bson.A{
		BsonD("$match", pipe),
		BsonD("$limit", 10),
	}
}
