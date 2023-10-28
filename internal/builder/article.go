package builder

import (
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

// list of paginated articles
func QrStrLookupArticle(cfg *types.QueryCtx) bson.A {
	sortField := cfg.Sort
	if sortField == "" {
		sortField = "created_at"
	}

	return bson.A{
		BsonD("$match", cfg.Search),
		BsonOperator("$sort", sortField, cfg.Order),
		BsonD("$skip", cfg.Skip),
		BsonD("$limit", cfg.Limit),
		QrStrLookupAccount("author"),
		BsonD("$unset", "author._id"),
		BsonOperator("$addFields", "author", BsonOperWithArray("$arrayElemAt", []interface{}{"$author", 0})),
		QrStrLookupAccount("co_authors"),
		BsonD("$unset", "co_authors._id"),
	}
}
