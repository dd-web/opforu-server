package builder

import "go.mongodb.org/mongo-driver/bson"

var (
	ASSET_SRC_PUBLIC_INCLUDE_FIELDS = []string{"details", "asset_type"}

	// this is the projection for after source lookup
	ASSET_PUBLIC_PROJECTION = map[string]string{
		"avatar":      "$source.details.avatar",
		"source":      "$source.details.source",
		"asset_type":  "$source.asset_type",
		"file_name":   "$file_name",
		"description": "$description",
		"tags":        "$tags",
		"created_at":  "$created_at",
		"updated_at":  "$updated_at",
	}
)

func QrStrLookupAssets() bson.D {
	pipe := bson.A{
		BsonLookup("asset_sources", "source_id", "_id", "source", bson.D{}, bson.A{}),
		BsonOperator("$addFields", "source", BsonOperWithArray("$arrayElemAt", []interface{}{"$source", 0})),
		BsonProjectionMap(ASSET_PUBLIC_PROJECTION),
		BsonOperWithArray("$unset", []interface{}{"_id"}),
	}
	return BsonLookup("assets", "assets", "_id", "assets", bson.D{}, pipe)
}
