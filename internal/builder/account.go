package builder

import (
	"go.mongodb.org/mongo-driver/bson"
)

var (
	public_account_fields = []string{"username", "role", "status", "created_at", "updated_at", "deleted_at"}
)

// account lookup pipeline
func QrStrLookupAccount(pk string) bson.D {
	projection := bson.A{BsonProjection(public_account_fields, 1)}
	return BsonLookup("accounts", pk, "_id", pk, bson.D{}, projection)
}
