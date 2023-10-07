package builder

import (
	"github.com/dd-web/opforu-server/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// returns an bson array of bson.D key/value pairs that together construct an aggregation query for
// mongo driver aggregation pipeline
// takes in the board's short name, page number and count per page
func QrStrPublicBoard(short string, cfg *utils.QueryConfig) bson.A {
	sortField := cfg.Sort
	if sortField == "" {
		sortField = "updated_at"
	}

	skip := (cfg.PageInfo.Current - 1) * cfg.PageInfo.PageSize
	return bson.A{
		BsonOperator("$match", "short", short),
		QrStrLookupThread(sortField, cfg.Order, skip, cfg.PageInfo.PageSize),
	}
}
