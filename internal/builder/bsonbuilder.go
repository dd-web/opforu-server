// bsonbuilder.go
//
// This package is to ease the construction of binary key value pair objects (bson)
// specifically, this is for use with the mongo-go-driver.
//
// Generally, a single key/value pair constructor will have a cooresponding func "BsonX" where X is the
// type of object to construct. Internally all key/value pairs are bson.E but it's not alway necessary
// to specify this when constructing the object.
//
// I specifically define Key and Value fields to please the linter and make it very clear what is going on.
// Again, generally this isn't necessary but it's good to see exactly what is going on.
//
// If an object is made up of multiple ground level bson.E fields, it's cooresponding func "ComposeBsonX" will
// take a map[string]any containing the key/value pairs to be constructed. The object will be constructed
// in the order of the map received, if it matters (bson.D)

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
	var d bson.D
	for k, v := range vals {
		d = append(d, BsonE(k, v))
	}
	return d
}

// Creates a $lookup bson obj with given parameters
func BsonLookup(col, pk, fk, as string, let bson.D, pipe bson.A) bson.D {
	lookupVal := ComposeBsonD(map[string]any{
		"from":         col,
		"localField":   pk,
		"foreignField": fk,
		"let":          let,
		"pipeline":     pipe,
		"as":           as,
	})

	return bson.D{{
		Key:   "$lookup",
		Value: lookupVal,
	}}
}

// Creates a $project bson obj with the given values as fields to include or exclude
// ie: {"name": 1, "age": 1} would pass ["name", "age"], 1 to include those fields or -1 to exclude
func BsonProjection(keys []string, val int) bson.D {
	valM := make(map[string]any)
	for _, v := range keys {
		valM[v] = val
	}
	projectVal := ComposeBsonD(valM)

	return bson.D{{
		Key:   "$project",
		Value: projectVal,
	}}
}

// creates a projection map from a map[string]T - for more specific control over the projection
func BsonProjectionMap[T comparable](vals map[string]T) bson.D {
	pjmap := make(map[string]T)
	for k, v := range vals {
		pjmap[k] = v
	}
	projection := ComposeBsonD(pjmap)

	return bson.D{{
		Key:   "$project",
		Value: projection,
	}}
}

// returns a full bson.D pipeline object for a $match or other operators
func BsonOperator(op string, k string, v any) bson.D {
	return BsonD(op, BsonD(k, v))
}

// same as BsonOperator but for operators that take an array instead of k/v pair
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

	return bson.A{
		match,
		BsonOperator("$sort", sort, cfg.Order),
		BsonD("$skip", cfg.Skip),
		BsonD("$limit", cfg.Limit),
	}, nil

}
