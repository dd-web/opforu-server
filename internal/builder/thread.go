package builder

import (
	"fmt"

	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// List of paginated thread previews for a board
func QrStrLookupThreads(boardID primitive.ObjectID, cfg *types.QueryCtx) (bson.A, error) {
	if boardID == primitive.NilObjectID {
		fmt.Println("Cannot lookup threads with invalid board")
		return nil, fmt.Errorf("invalid board")
	}

	innerMatch := bson.D{{Key: "board", Value: boardID}}
	innerMatch = append(innerMatch, cfg.Search...)

	sortField := cfg.Sort
	if sortField == "" {
		sortField = "updated_at"
	}

	return bson.A{
		BsonD("$match", innerMatch),
		BsonOperator("$sort", sortField, cfg.Order),
		BsonD("$skip", cfg.Skip),
		BsonD("$limit", cfg.Limit),
		BsonOperator("$addFields", "post_count", BsonD("$size", "$posts")),
		QrStrLookupPosts("post_number", -1, 5),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		QrStrLookupIdentity("mods"),
		BsonOperWithArray("$unset", []interface{}{"board", "account"}),
		// lookup media here
	}, nil
}

// a single thread with all posts populated
func QrStrEntireThread(slug string, cfg *types.QueryCtx) bson.A {
	return bson.A{
		BsonOperator("$match", "slug", slug),
		QrStrLookupPosts("post_number", 1, 0),
		QrStrLookupIdentity("creator"),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		QrStrLookupIdentity("mods"),
		BsonOperWithArray("$unset", []interface{}{"board", "account"}),
		// lookup media here
	}
}

// bson.A{
// 	bson.D{{"$match", bson.D{{"short", acronym}}}},
// 	bson.D{{"$lookup", bson.D{
// 				{"from", "threads"},
// 				{"localField", "threads"},
// 				{"foreignField", "_id"},
// 				{"let", bson.D{{"board", "$_id"}}},
// 				{"pipeline",
// 					bson.A{
// 						bson.D{{"$addFields", bson.D{{"post_count", bson.D{{"$size", "$posts"}}}}}},
// 						bson.D{{"$sort", bson.D{{"updated_at", -1}}}},
// 						bson.D{{"$skip", sk}},
// 						bson.D{{"$limit", perPage}},
// 						bson.D{
// 							{"$lookup",
// 								bson.D{
// 									{"from", "posts"},
// 									{"localField", "posts"},
// 									{"foreignField", "_id"},
// 									{"pipeline",
// 										bson.A{
// 											bson.D{{"$sort", bson.D{{"post_number", 1}}}},
// 											bson.D{{"$limit", 5}},
// 											bson.D{
// 												{"$lookup",
// 													bson.D{
// 														{"from", "media"},
// 														{"localField", "media"},
// 														{"foreignField", "_id"},
// 														{"pipeline",
// 															bson.A{
// 																bson.D{
// 																	{"$unset",
// 																		bson.A{
// 																			"source",
// 																		},
// 																	},
// 																},
// 															},
// 														},
// 														{"as", "media"},
// 													},
// 												},
// 											},
// 											bson.D{
// 												{"$lookup",
// 													bson.D{
// 														{"from", "identities"},
// 														{"localField", "creator"},
// 														{"foreignField", "_id"},
// 														{"pipeline",
// 															bson.A{
// 																bson.D{
// 																	{"$project",
// 																		bson.D{
// 																			{"name", 1},
// 																			{"style", 1},
// 																			{"role", 1},
// 																			{"status", 1},
// 																			{"created_at", 1},
// 																			{"updated_at", 1},
// 																		},
// 																	},
// 																},
// 															},
// 														},
// 														{"as", "creator"},
// 													},
// 												},
// 											},
// 											bson.D{
// 												{"$addFields",
// 													bson.D{
// 														{"creator",
// 															bson.D{
// 																{"$arrayElemAt",
// 																	bson.A{
// 																		"$creator",
// 																		0,
// 																	},
// 																},
// 															},
// 														},
// 													},
// 												},
// 											},
// 											bson.D{
// 												{"$unset",
// 													bson.A{
// 														"thread",
// 														"board",
// 													},
// 												},
// 											},
// 										},
// 									},
// 									{"as", "posts"},
// 								},
// 							},
// 						},
// 						bson.D{
// 							{"$lookup",
// 								bson.D{
// 									{"from", "identities"},
// 									{"localField", "creator"},
// 									{"foreignField", "_id"},
// 									{"pipeline",
// 										bson.A{
// 											bson.D{
// 												{"$project",
// 													bson.D{
// 														{"name", 1},
// 														{"style", 1},
// 														{"role", 1},
// 														{"status", 1},
// 														{"created_at", 1},
// 														{"updated_at", 1},
// 													},
// 												},
// 											},
// 										},
// 									},
// 									{"as", "creator"},
// 								},
// 							},
// 						},
// 						bson.D{
// 							{"$addFields",
// 								bson.D{
// 									{"creator",
// 										bson.D{
// 											{"$arrayElemAt",
// 												bson.A{
// 													"$creator",
// 													0,
// 												},
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 						bson.D{
// 							{"$lookup",
// 								bson.D{
// 									{"from", "identities"},
// 									{"localField", "mods"},
// 									{"foreignField", "_id"},
// 									{"pipeline",
// 										bson.A{
// 											bson.D{
// 												{"$project",
// 													bson.D{
// 														{"name", 1},
// 														{"style", 1},
// 														{"role", 1},
// 														{"status", 1},
// 														{"created_at", 1},
// 														{"updated_at", 1},
// 													},
// 												},
// 											},
// 										},
// 									},
// 									{"as", "mods"},
// 								},
// 							},
// 						},
// 						bson.D{
// 							{"$lookup",
// 								bson.D{
// 									{"from", "media"},
// 									{"localField", "media"},
// 									{"foreignField", "_id"},
// 									{"pipeline",
// 										bson.A{
// 											bson.D{
// 												{"$unset",
// 													bson.A{
// 														"source",
// 													},
// 												},
// 											},
// 										},
// 									},
// 									{"as", "media"},
// 								},
// 							},
// 						},
// 						bson.D{
// 							{"$unset",
// 								bson.A{
// 									"board",
// 								},
// 							},
// 						},
// 					},
// 				},
// 				{"as", "threads"},
// 			},
// 		},
// 	},
// }
