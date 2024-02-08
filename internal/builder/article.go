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
		QrStrLookupArticleAuthor("author"),
		BsonD("$unset", "author.author._id"),
		BsonOperator("$addFields", "author", BsonOperWithArray("$arrayElemAt", []interface{}{"$author", 0})),
		BsonOperator("$addFields", "author", "$author.author"),
		QrStrLookupArticleAuthor("co_authors"),
		BsonOperator("$addFields", "co_authors", "$co_authors.author"),
		BsonD("$unset", "co_authors._id"),
		QrStrLookupAssets(),
	}
}

// article author lookup
func QrStrLookupArticleAuthor(pk string) bson.D {
	return BsonLookup("article_authors", pk, "_id", pk, bson.D{}, getAuthorPipe())
}

// author lookup internal pipe - internal processing for indirect author lists (intermediate author data struct)
func getAuthorPipe() bson.A {
	return bson.A{
		QrStrLookupAccount("author"),
		BsonOperator("$addFields", "author", BsonOperWithArray("$arrayElemAt", []interface{}{"$author", 0})),
		BsonOperator("$addFields", "author.username", BsonOperWithArray("$cond", []interface{}{"$anonymize", "anonymous", "$author.username"})),
	}
}
