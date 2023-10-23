package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	Client *mongo.Client
	DB     *mongo.Database
	Name   string

	BoardIDs map[string]primitive.ObjectID

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
		BoardIDs:  map[string]primitive.ObjectID{},
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

// Runs the passed aggegation pipeline on the specified collection/column and returns the results
// results must be a slice because bson.M is not a comparable type
func (s *Store) RunAggregation(col string, pipe any) ([]bson.M, error) {
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

// fetches and stores frequently used board data in memory for quicker access
func (s *Store) HydrateBoardIDs() error {
	collection := s.DB.Collection("boards")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}

	defer func() {
		cursor.Close(ctx)
	}()

	for cursor.Next(ctx) {
		var board types.Board
		err := cursor.Decode(&board)
		if err != nil {
			fmt.Println("Error decoding board", err)
			continue
		}
		s.BoardIDs[board.Short] = board.ID
	}

	return nil
}

// Finds a single Board document by it's short name, unmarshals it into a Board struct and returns a pointer to it
func (s *Store) FindBoardByShort(short string) (*types.Board, error) {
	collection := s.DB.Collection("boards")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result types.Board
	err := collection.FindOne(ctx, bson.D{{Key: "short", Value: short}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	result.Threads = []primitive.ObjectID{}

	return &result, nil
}

// Finds the total number of threads matching the given filter
func (s *Store) CountThreadMatch(boardId primitive.ObjectID, filter bson.D) (int64, error) {
	countOpts := options.Count().SetMaxTime(5 * time.Second)
	collection := s.DB.Collection("threads")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	countFilter := append(bson.D{{Key: "board", Value: boardId}}, filter...)

	count, err := collection.CountDocuments(ctx, countFilter, countOpts)
	if err != nil {
		fmt.Println("Error getting total record count", err)
		return 0, err
	}

	fmt.Println("Total matching records:", count)

	return count, nil
}

// find a session by it's session_id
func (s *Store) FindSession(session string) (*types.Session, error) {
	collection := s.DB.Collection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result types.Session
	err := collection.FindOne(ctx, bson.D{{Key: "session_id", Value: session}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	fmt.Println("Matching Session:", result)

	return &result, nil
}

// find an account by it's _id
func (s *Store) FindAccountByID(id primitive.ObjectID) (*types.Account, error) {
	collection := s.DB.Collection("accounts")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result types.Account
	err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// fins an account by it's username or email address
// if email is empty string username will be supplied for both parameters
func (s *Store) FindAccountByUsernameOrEmail(username string, email string) (*types.Account, error) {
	collection := s.DB.Collection("accounts")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var criteraTwo string = email
	if email == "" {
		criteraTwo = username
	}

	var result types.Account
	err := collection.FindOne(ctx, bson.D{{Key: "$or", Value: bson.A{bson.D{{Key: "username", Value: username}}, bson.D{{Key: "email", Value: criteraTwo}}}}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// find an active session by it's account id
func (s *Store) FindSessionFromUser(act primitive.ObjectID) (*types.Session, error) {
	collection := s.DB.Collection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result types.Session
	err := collection.FindOne(ctx, bson.D{{Key: "account", Value: act}, {Key: "active", Value: true}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
