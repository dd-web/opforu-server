package builder

import (
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

// returns a pipeline object
// pk is the field name to reference the lookup from
func QrStrLookupAccount(pk string) bson.D {
	projection := bson.A{BsonProjection(types.PUBLIC_ACCOUNT_FIELDS, 1)}
	return BsonLookup("accounts", pk, "_id", pk, bson.D{}, projection)
}
