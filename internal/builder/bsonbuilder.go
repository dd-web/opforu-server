// bsonbuilder.go
//
// This package is to ease the construction of binary k/v objects
// specifically, this is for use with the mongo-driver.

package builder

import (
	"fmt"

	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create a bson.A - takes an array of any type, each will be inserted into the bson.A as itself
func BsonA(vals []any) bson.A {
	arr := bson.A{}
	for _, k := range vals {
		arr = append(arr, k)
	}
	return arr
}

// Creates a bson.E with the given key/value pair
func BsonE(k string, v any) bson.E {
	return bson.E{
		Key:   k,
		Value: v,
	}
}

// Creates a bson.D with a single key/value pair of bson.E values
func BsonD(k string, v any) bson.D {
	return bson.D{
		BsonE(k, v),
	}
}

// Creates a bson.D including the given key/value pair map as bson.E values
func ComposeBsonD[T comparable](vals map[string]T) bson.D {
	d := bson.D{}
	for k, v := range vals {
		d = append(d, BsonE(k, v))
	}
	return d
}

// Creates a $lookup bson obj with given parameters
func BsonLookup(col, pk, fk, as string, let bson.D, pipe bson.A) bson.D {
	return BsonD("$lookup", ComposeBsonD[any](map[string]any{
		"from":         col,
		"localField":   pk,
		"foreignField": fk,
		"let":          let,
		"pipeline":     pipe,
		"as":           as,
	}))
}

type BSONProjectValue int

const (
	BSONProjectInclude BSONProjectValue = 1
	BSONProjectExclude BSONProjectValue = -1
)

// Creates a bson projection with keys either included or excluded
// with inclusion, exclusions are explicit. the inverse is true of exclusions
func BsonProjection(keys []string, val BSONProjectValue) bson.D {
	valM := make(map[string]any)
	for _, v := range keys {
		valM[v] = val
	}
	return BsonD("$project", ComposeBsonD[any](valM))
}

// creates a projection map from a map[string]T - for more specific control over the projection
func BsonProjectionMap[T comparable](vals map[string]T) bson.D {
	pjmap := make(map[string]T)
	for k, v := range vals {
		pjmap[k] = v
	}
	return BsonD("$project", ComposeBsonD[T](pjmap))
}

// creates bson for pipeline operators
func BsonOperator(op string, k string, v any) bson.D {
	return BsonD(op, BsonD(k, v))
}

// same as BsonOperator but for operators that take an array instead of bson
func BsonOperWithArray(op string, v []any) bson.D {
	return BsonD(op, BsonA(v))
}

// starts a paginated pipeline with a match on the given key/value pair
func StartPaginatedPipe(mkey string, mval primitive.ObjectID, cfg *types.QueryCtx) (bson.A, error) {
	if mval == primitive.NilObjectID {
		return nil, fmt.Errorf("invalid match key: %s", mkey)
	}

	match := BsonD("$match", BsonD(mkey, mval))
	match = append(match, cfg.Search...)

	sort := cfg.Sort
	if sort == "" {
		sort = "updated_at"
	}

	return BsonA([]any{
		match,
		BsonOperator("$sort", sort, cfg.Order),
		BsonD("$skip", cfg.Skip),
		BsonD("$limit", cfg.Limit),
	}), nil
}
