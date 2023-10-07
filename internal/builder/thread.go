package builder

import "go.mongodb.org/mongo-driver/bson"

func QrStrLookupThread(sortBy string, sortDir, skip, limit int) bson.D {
	pipe := bson.A{
		BsonOperator("$addFields", "post_count", BsonD("$size", "$posts")),
		BsonOperator("$sort", sortBy, sortDir),
		BsonD("$skip", skip),
		BsonD("$limit", limit),
		QrStrLookupPosts("post_number", 1, 5),
		QrStrLookupIdentityCreator(),
		BsonOperator("$addFields", "creator", BsonOperWithArray("$arrayElemAt", []interface{}{"$creator", 0})),
		QrStrLookupIdentityMods(),
		// lookup media here
		BsonOperWithArray("$unset", []interface{}{"board"}),
	}

	return BsonLookup("threads", "threads", "_id", "threads", BsonD("board", "$_id"), pipe)
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
