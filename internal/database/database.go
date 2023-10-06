package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	Client *mongo.Client
	DB     *mongo.Database
	Name   string

	StartedAt time.Time
	EndedAt   time.Time
}

// creates a new store with a connection to the supplied database name
func NewStore(name string) (*Store, error) {
	var ended time.Time
	uri := parseURIFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		ended = time.Now().UTC()
		cancel()
	}()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("Error connecting to MongoDB")
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
		return nil, err
	}

	db := client.Database(name)

	return &Store{
		Client:    client,
		DB:        db,
		Name:      name,
		StartedAt: time.Now().UTC(),
		EndedAt:   ended,
	}, nil
}

// constructs a connection string from env vars
func parseURIFromEnv() string {
	if os.Getenv("MONGO_URI") != "" {
		return os.Getenv("MONGO_URI")
	}

	return fmt.Sprintf("mongodb://%s:%s/",
		assertEnvStr(os.Getenv("DB_HOST")),
		assertEnvStr(os.Getenv("DB_PORT")),
	)
}

// ensures the required string has a value
func assertEnvStr(v string) string {
	if v == "" {
		log.Fatal("Invalid Environemtn Variable")
	}
	return v
}

// Constructs and runs and aggregation query on the specified collection/column using the supplied query config
// always returns a slice, if looking for a single result use the first element
// cannot be any type because primitive.M is not a comparable type
func (s *Store) RunAggregation(col string, q utils.QueryConfig) ([]bson.M, error) {
	collection := s.DB.Collection(col)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// @TODO: binary object constructor. ABSOLUTELY NECESARY - MUST be used for queries
	// for now using a stand in bson.A to get it working
	// bs := bson.A{
	// 	bson.D{{
	// 		Key:   "$match",
	// 		Value: bson.D{{Key: "short", Value: "g"}},
	// 	}},
	// }
	matcher := builder.BsonD("$match", "short", "g")
	bs := bson.A{matcher}

	cursor, err := collection.Aggregate(ctx, bs)
	if err != nil {
		return nil, err
	}

	defer func() {
		cursor.Close(ctx)
	}()

	records := []bson.M{}

	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// Persists multiple new documents to a specified collection/column
func (s *Store) SaveNewMulti(documents []any, col string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.DB.Collection(col)
	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully saved %d documents to %s column\n", len(documents), col)
	return nil
}

// Persists a single new document to a specified collection/column
func (s *Store) SaveNewSingle(document any, col string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.DB.Collection(col)
	_, err := collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully saved document to %s column\n", col)
	return nil
}
