package builder

import (
	"go.mongodb.org/mongo-driver/bson"
)

var (
	// permissions
	public_account_fields   = []string{"username", "role", "status", "created_at", "updated_at", "deleted_at"}
	personal_account_fields = []string{"email", "_id"}
)

// returns a pipeline object
// pk is the field name to reference the lookup from
func QrStrLookupAccount(pk string) bson.D {
	projection := bson.A{BsonProjection(public_account_fields, 1)}
	return BsonLookup("accounts", pk, "_id", pk, bson.D{}, projection)
}
