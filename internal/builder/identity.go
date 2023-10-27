package builder

import (
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

// returns a pipeline object that can be used as an internal pipeline for a $lookup or other aggregate
func QrStrLookupIdentityCreator() bson.D {
	projection := bson.A{BsonProjection(types.PUBLIC_IDENTITY_FIELDS, 1)}
	return BsonLookup("identities", "creator", "_id", "creator", bson.D{}, projection)
}

// returns a pipeline object that can be used as an internal pipeline for a $lookup or other aggregate
func QrStrLookupIdentityMods() bson.D {
	projection := bson.A{BsonProjection(types.PUBLIC_IDENTITY_FIELDS, 1)}
	return BsonLookup("identities", "mods", "_id", "mods", bson.D{}, projection)
}

// query string for counting unique posters within a thread (partial agg pipeline, meant for internal pipeline usage)
func QrStrGroupUniquePosters() bson.D {
	pct := bson.D{{Key: "_id", Value: "$creator"}, {Key: "postCount", Value: BsonD("$count", nil)}}
	return BsonD("$group", pct)
}

//groups posts by identity, this way we can get the number of unique posters in a thread
// use this on the post collection after it's been aggregated
//  $group: {
// 	_id: "$creator",
// 	postCount: {
// 		$count: {}
// 	}
// }
