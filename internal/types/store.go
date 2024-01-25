package types

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dd-web/opforu-server/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Server Cache
// high level cache for extremely frequently used data
// TODO: implement refreshes & invalidation
type ServerCache struct {
	Boards   map[string]*Board               // short -> Board
	Sessions map[string]*Session             // session_id -> Session
	Accounts map[primitive.ObjectID]*Account // _id -> Account

	StartedAt *time.Time
	EndedAt   *time.Time
}

// New Server Cache
// creates a new server cache with empty maps and the current time
func NewServerCache() *ServerCache {
	ts := time.Now().UTC()
	return &ServerCache{
		Boards:    map[string]*Board{},
		Sessions:  map[string]*Session{},
		Accounts:  map[primitive.ObjectID]*Account{},
		StartedAt: &ts,
	}
}

type Store struct {
	Client *mongo.Client
	DB     *mongo.Database
	Name   string

	Cache *ServerCache // high level cache for frequently used data

	StartedAt *time.Time
	EndedAt   *time.Time
}

// New Store
// creates a new store for database operations
func NewStore(dbname string) (*Store, error) {
	var ended time.Time
	uri := utils.ParseURIFromEnv()
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

	db := client.Database(dbname)
	ts := time.Now().UTC()

	return &Store{
		Client:    client,
		DB:        db,
		Name:      dbname,
		StartedAt: &ts,
		EndedAt:   &ended,
		Cache:     NewServerCache(),
	}, nil
}

// Run Aggregation
// - accepts a string of the collection name
// - accepts a (usually binary object notation) pipeline to be ran
// - returns a slice of bson.M containing the results
// - returns an error if one occurs
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

// Save Multiple Documents
// - accepts a slice of (usually binary object notations) documents but could be other types
// - accepts a string of the collection name
// - returns an error if one occurs
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

// Save a Single Document
// - accepts a (usually binary object notation) document but could be other types
// - accepts a string of the collection name
// - returns an error if one occurs
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

// Delete a single Document
// - accepts a primitive.ObjectID of the document to be deleted
// - accepts a string of the collection name
// - returns an error if one occurs
func (s *Store) DeleteSingle(id primitive.ObjectID, col string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.DB.Collection(col)
	_, err := collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}

	fmt.Printf("Successfully deleted document from %s column\n", col)
	return nil
}

// Hydrate Cache
// - returns an error if one occurs
//
//	Attempts to hydrate the cache with frequently used data from the database, like boards.
func (s *Store) HydrateCache() error {
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
		var board Board
		err := cursor.Decode(&board)
		if err != nil {
			fmt.Println("Error decoding board", err)
			continue
		}
		s.Cache.Boards[board.Short] = &board
	}

	return nil
}

// Find Board By Short Name
// - accepts a string of the board short name
// - returns a pointer to the board
//
//	Attempts to find the board in the cache first, if it's not found it will be looked up in the database and cached for future use.
func (s *Store) FindBoardByShort(short string) (*Board, error) {
	var board *Board
	c_board, ok := s.Cache.Boards[short]
	if ok {
		board = c_board
	}

	if board == nil {
		collection := s.DB.Collection("boards")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := collection.FindOne(ctx, bson.D{{Key: "short", Value: short}}).Decode(&board)
		if err != nil {
			return nil, err
		}
		s.Cache.Boards[short] = board
	}
	return board, nil
}

// Count Results
// - accepts a string of the collection name
// - accepts a bson.D of the filter
// - returns an int64 of the count
//
//	Counts the number of documents matching the given filter in the specified collection.
//	useful for pagination since when we query for results we're only receiving a subset of the total results.
func (s *Store) CountResults(col string, filter bson.D) int64 {
	var result int64 = 0

	count_options := options.Count().SetMaxTime(5 * time.Second)
	collection := s.DB.Collection(col)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, _ = collection.CountDocuments(ctx, filter, count_options)
	return result
}

// Find Session
// - accepts a string of the session id
// - returns a pointer to the session
// - returns an error if one occurs
//
//	Attempts to find the session in the cache first, if it's not found it will be looked up in the database and cached for future use.
func (s *Store) FindSession(id string) (*Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session id is empty")
	}

	var session *Session

	c_session, ok := s.Cache.Sessions[id]
	if ok {
		session = c_session
	}

	if session == nil {
		collection := s.DB.Collection("sessions")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := collection.FindOne(ctx, bson.D{{Key: "session_id", Value: id}}).Decode(&session)
		if err != nil {
			fmt.Println("FindSession error:", err)
			return nil, err
		}
		s.Cache.Sessions[id] = session
	}

	return session, nil
}

// Find Account By Username or Email
// - accepts a string of the username
// - accepts a string of the email (optional - can be empty string)
// - returns a pointer to the account
// - returns an error if one occurs
//
//	Always queries the database since we don't index by username or email in the cache and it's not a frequently used query.
//	Email is optional, if it's empty it will use the username (first parameter) to search by both username and email.
func (s *Store) FindAccountByUsernameOrEmail(username string, email string) (*Account, error) {
	collection := s.DB.Collection("accounts")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var criteraTwo string = email
	if email == "" {
		criteraTwo = username
	}

	var result Account
	err := collection.FindOne(ctx, bson.D{{Key: "$or", Value: bson.A{bson.D{{Key: "username", Value: username}}, bson.D{{Key: "email", Value: criteraTwo}}}}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Find Account By Session ID
// - accepts a string of the session id
// - returns a pointer to the associated account
// - returns an error if one occurs
//
// attempts to look up both account and session from the cache first, if unavailable they will be looked up in the database and cached for future use.
func (s *Store) FindAccountFromSession(id string) (*Account, error) {
	if id == "" {
		return nil, fmt.Errorf("session id is empty")
	}

	var session *Session
	var account *Account

	c_session, ok := s.Cache.Sessions[id]
	if ok {
		session = c_session
		c_account, ok := s.Cache.Accounts[session.AccountID]
		if ok {
			account = c_account
		}
	}

	if account != nil {
		return account, nil
	}

	if session == nil { // it's possible session is cached but not the account - check it
		collection := s.DB.Collection("sessions")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := collection.FindOne(ctx, bson.D{{Key: "session_id", Value: id}}).Decode(&session)
		if err != nil {
			return nil, err
		}
		session.IsExpiringSoon()

		s.Cache.Sessions[id] = session
	}

	collection := s.DB.Collection("accounts")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: session.AccountID}}).Decode(&account)
	if err != nil {
		return nil, err
	}

	s.Cache.Accounts[session.AccountID] = account

	return account, nil
}

// Find Asset Source with a hash collisions
// - accepts a byte slice of the hash
// - accepts a types.HashMethod of the hash method used
// - returns a pointer to the asset source
// - returns an error if one occurs
//
// still lots to do here. just getting something working for now.
func (s *Store) AssetHashCollision(hash []byte, method HashMethod) (*AssetSource, error) {
	collection := s.DB.Collection("asset_sources")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result AssetSource

	err := collection.FindOne(ctx, bson.D{{Key: "details.source.hash_md5", Value: hash}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
