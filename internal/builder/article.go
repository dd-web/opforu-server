package builder

import (
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

// list of paginated articles
func QrStrLookupArticleList(cfg *types.QueryCtx) bson.A {
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
		QrStrLookupAssets("assets"),
	}
}

// single article with populated comments
func QrStrLookupArticle(slug string) bson.A {
	return bson.A{
		BsonOperator("$match", "slug", slug),
		QrStrLookupArticleAuthor("author"),
		BsonD("$unset", "author.author._id"),
		BsonOperator("$addFields", "author", BsonOperWithArray("$arrayElemAt", []interface{}{"$author", 0})),
		BsonOperator("$addFields", "author", "$author.author"),
		BsonD("$unset", "author.created_at"),
		BsonD("$unset", "author.updated_at"),
		QrStrLookupArticleAuthor("co_authors"),
		BsonOperator("$addFields", "co_authors", "$co_authors.author"),
		BsonD("$unset", "co_authors._id"),
		BsonD("$unset", "co_authors.created_at"),
		BsonD("$unset", "co_authors.updated_at"),
		BsonLookup("article_comments", "comments", "_id", "comments", bson.D{}, getCommentPipe()),
		QrStrLookupAssets("assets"),
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

// comment lookup internal pipe
func getCommentPipe() bson.A {
	return bson.A{
		QrStrLookupAccount("author"),
		BsonOperator("$addFields", "author", BsonOperWithArray("$arrayElemAt", []interface{}{"$author", 0})),
		BsonOperator("$addFields", "author.username", BsonOperWithArray("$cond", []interface{}{"$author_anonymous", "anonymous", "$author.username"})),
		BsonD("$unset", "author._id"),
		BsonD("$unset", "author.created_at"),
		BsonD("$unset", "author.updated_at"),
		BsonD("$unset", "author_anonymous"),
		QrStrLookupAssets("assets"),
	}
}
