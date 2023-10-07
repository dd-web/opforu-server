package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
func (s *Store) RunAggregation(col string, pipe interface{}) ([]bson.M, error) {
	collection := s.DB.Collection(col)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}

	defer func() {
		cursor.Close(ctx)
	}()

	var records []bson.M
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

// Uses the query config object to find documents in the specified collection/column
// returns a slice of bson.M, if it should be returned immediately it doesn't need to
// be decoded into a struct and can be returned as is into the json response.
// otherwise each model has an UnMarshal method that decodes it into a struct of that type.
func (s *Store) Find(col string, cfg *utils.QueryConfig) ([]bson.M, error) {
	matchCount, err := s.TotalRecordCount(col, cfg.Filter)
	if err != nil {
		fmt.Println("Error getting total record count", err)
	}

	if cfg.PageInfo != nil {
		cfg.PageInfo.Update(int(matchCount))
	}

	collection := s.DB.Collection(col)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOpts := options.Find()
	if cfg != nil {
		findOpts.SetLimit(cfg.Limit)
		findOpts.SetSkip(cfg.Skip)
		findOpts.SetSort(bson.M{cfg.Sort: cfg.Order})
	}

	var results []bson.M = []bson.M{}

	cursor, err := collection.Find(ctx, cfg.Filter, findOpts)
	if err != nil {
		return nil, err
	}

	defer func() {
		cursor.Close(ctx)
	}()

	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			fmt.Println("Error decoding result", err)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// get the total number of records in a collection
func (s *Store) TotalRecordCount(col string, filter interface{}) (int64, error) {
	countOpts := options.Count().SetMaxTime(2 * time.Second)
	collection := s.DB.Collection(col)

	count, err := collection.CountDocuments(context.Background(), filter, countOpts)
	if err != nil {
		fmt.Println("Error getting total record count", err)
		return 0, err
	}

	fmt.Println("Total record count:", count)

	return count, nil
}
