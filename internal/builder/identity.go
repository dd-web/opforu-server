package builder

import "go.mongodb.org/mongo-driver/bson"

var (
	publicIdentityFields = []string{"name", "style", "role", "status", "created_at", "updated_at", "deleted_at"}
)

// returns a pipeline object that can be used as an internal pipeline for a $lookup or other aggregate
func QrStrLookupIdentityCreator() bson.D {
	projection := bson.A{BsonProjection(publicIdentityFields, 1)}
	return BsonLookup("identities", "creator", "_id", "creator", bson.D{}, projection)
}

// returns a pipeline object that can be used as an internal pipeline for a $lookup or other aggregate
func QrStrLookupIdentityMods() bson.D {
	projection := bson.A{BsonProjection(publicIdentityFields, 1)}
	return BsonLookup("identities", "mods", "_id", "mods", bson.D{}, projection)
}

//groups posts by identity, this way we can get the number of unique posters in a thread
// use this on the post collection after it's been aggregated
//  $group: {
// 	_id: "$creator",
// 	postCount: {
// 		$count: {}
// 	}
// }
//