package builder

import (
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

// returns a pipeline object that can be used as an internal pipeline for a $lookup or other aggregate
// pk is the field name to reference the lookup from
func QrStrLookupIdentity(pk string) bson.D {
	projection := bson.A{BsonProjection(types.PUBLIC_IDENTITY_FIELDS, 1)}
	return BsonLookup("identities", pk, "_id", pk, bson.D{}, projection)
}
