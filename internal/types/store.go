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

// High level server cache for frequently used data to avoid network calls
//
// Boards are cached on server start, sessions and accounts are cached as they're accessed.
type ServerCache struct {
	Boards    map[string]*Board               // short -> Board
	Sessions  map[string]*Session             // session_id -> Session
	Accounts  map[primitive.ObjectID]*Account // _id -> Account
	StartedAt *time.Time
	EndedAt   *time.Time
}

func NewServerCache() *ServerCache {
	ts := time.Now().UTC()
	return &ServerCache{
		Boards:    map[string]*Board{},
		Sessions:  map[string]*Session{},
		Accounts:  map[primitive.ObjectID]*Account{},
		StartedAt: &ts,
	}
}

// Global data store initialized on server start
//
// Store is initialized on server start and is referenced in all handlers.
// The RequestCtx for any particular handler will be updated with a reference
// to the store after it's called.
//
// This is all to avoid circular dependencies while giving store access essentially everywhere.
type Store struct {
	Name      string
	Client    *mongo.Client
	DB        *mongo.Database
	Cache     *ServerCache
	StartedAt *time.Time
	EndedAt   *time.Time
}

func NewStore(dbname string) (*Store, error) {
	// var ended time.Time
	ended := time.Time{}
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

/*******************************************************************************************
 * Article Operations
 *******************************************************************************************/

// Find article by slug
// - accepts a string of the articles slug
// - returns pointer to the article
func (s *Store) FindArticleBySlug(slug string) (*Article, error) {
	article := &Article{}

	collection := s.DB.Collection("articles")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.D{{Key: "slug", Value: slug}}).Decode(&article)
	if err != nil {
		return nil, err
	}

	return article, nil
}

// Update the provided article
// uses the provided article's ID to determine which to update
// - accepts a pointer to an article
// - returns an error if one occurred, else nil
func (s *Store) UpdateArticle(article *Article) error {
	collection := s.DB.Collection("articles")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "_id", Value: article.ID}}

	_, err := collection.ReplaceOne(ctx, filter, article)
	if err != nil {
		return err
	}

	return nil
}

/*******************************************************************************************
 * Board Operations
 *******************************************************************************************/

// Find board by short
// - accepts a string of the board short name
// - returns a pointer to the board
func (s *Store) FindBoardByShort(short string) (*Board, error) {
	board := &Board{}

	collection := s.DB.Collection("boards")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.D{{Key: "short", Value: short}}).Decode(&board)
	if err != nil {
		return nil, err
	}

	return board, nil
}

// Find board by _id
// - accepts primitive.ObjectID of the board (_id)
// - returns a pointer to the board
func (s *Store) FindBoardByObjectID(id primitive.ObjectID) (*Board, error) {
	// var board *Board
	board := &Board{}

	collection := s.DB.Collection("boards")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&board)
	if err != nil {
		return nil, err
	}

	return board, nil
}

// Update the provided board
// uses the passed board's ID and short name to determine which to update. essentially replaces old with new.
// - accepts a pointer to a Board object
// - returns an error if one occurred, else nil
func (s *Store) UpdateBoard(board *Board) error {
	collection := s.DB.Collection("boards")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "_id", Value: board.ID}, {Key: "short", Value: board.Short}}

	_, err := collection.ReplaceOne(ctx, filter, board)
	if err != nil {
		return err
	}

	return nil
}

/*******************************************************************************************
 * Thread Operations
 *******************************************************************************************/

// Find thread by slug
// - accepts a string of thread slug
// - returns a pointer to the thread
func (s *Store) FindThreadBySlug(slug string) (*Thread, error) {
	// var thread *Thread
	thread := &Thread{}

	collection := s.DB.Collection("threads")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.D{{Key: "slug", Value: slug}}).Decode(&thread)
	if err != nil {
		return nil, err
	}

	return thread, nil
}

// Update the provided thread
// uses the passed thread's ID and slug to determine which to update, essentially replaces old with new.
// - accepts a pointer to a Thread object
// - returns an error if one occurred, else nil
func (s *Store) UpdateThread(thread *Thread) error {
	collection := s.DB.Collection("threads")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "_id", Value: thread.ID}, {Key: "slug", Value: thread.Slug}}

	_, err := collection.ReplaceOne(ctx, filter, thread)
	if err != nil {
		return err
	}

	return nil
}

/*******************************************************************************************
 * Account Operations
 *******************************************************************************************/

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

	criteraTwo := email
	if email == "" {
		criteraTwo = username
	}

	result := &Account{}
	err := collection.FindOne(ctx, bson.D{{Key: "$or", Value: bson.A{bson.D{{Key: "username", Value: username}}, bson.D{{Key: "email", Value: criteraTwo}}}}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Find Account By Session ID
// - accepts a string of the session id
// - returns a pointer to the associated account
// - returns an error if one occurs
func (s *Store) FindAccountFromSession(id string) (*Account, error) {
	if id == "" {
		return nil, fmt.Errorf("session id is empty")
	}

	session := &Session{}
	account := &Account{}

	{
		collection := s.DB.Collection("sessions")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := collection.FindOne(ctx, bson.D{{Key: "session_id", Value: id}}).Decode(&session)
		if err != nil {
			return nil, err
		}
	}

	{
		collection := s.DB.Collection("accounts")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: session.AccountID}}).Decode(&account)
		if err != nil {
			return nil, err
		}
	}

	return account, nil
}

/*******************************************************************************************
 * Session Operations
 *******************************************************************************************/

// Find Session
// - accepts a string of the session id
// - returns a pointer to the session
// - returns an error if one occurs
func (s *Store) FindSession(id string) (*Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session id is empty")
	}

	var session *Session

	collection := s.DB.Collection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.D{{Key: "session_id", Value: id}}).Decode(&session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

/*******************************************************************************************
 * Asset Operations
 *******************************************************************************************/

// runs a list of queries provided (hash collision query strings from builder)
// returns an AssetSource of the first matching hash collision
func (s *Store) AssetHashCollisionResolver(queries ...primitive.D) (*AssetSource, error) {
	// var found *AssetSource
	found := &AssetSource{}
	matched := false

	for _, query := range queries {
		result, err := s.assetColliderQuery(query)
		if err != nil {
			continue
		}
		if result != nil {
			found = result
			matched = true
			break
		}
	}

	if !matched {
		return nil, nil
	}

	return found, nil
}

// single query runner for the resolver
func (s *Store) assetColliderQuery(query primitive.D) (*AssetSource, error) {
	collection := s.DB.Collection("asset_sources")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := &AssetSource{
		ID: primitive.NilObjectID,
	}

	err := collection.FindOne(ctx, query).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.ID == primitive.NilObjectID {
		return nil, nil
	}

	return result, nil
}

// find asset (not source) by it's id
func (s *Store) FindAssetByID(id primitive.ObjectID) (*Asset, error) {
	collection := s.DB.Collection("assets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := &Asset{
		ID: primitive.NilObjectID,
	}

	err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// find asset source by it's hash (sha256)
func (s *Store) FindAssetSourceByHash(hash string) (*AssetSource, error) {
	collection := s.DB.Collection("asset_sources")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := &AssetSource{
		ID: primitive.NilObjectID,
	}

	err := collection.FindOne(ctx, bson.D{{Key: "details.source.hash_sha256", Value: hash}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.ID == primitive.NilObjectID {
		return nil, nil
	}

	return result, nil
}

// find all assets with given source id and account id
func (s *Store) FindAssetsBySourceIDAccountID(source_id, account_id primitive.ObjectID) ([]*Asset, error) {
	collection := s.DB.Collection("assets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{{Key: "source_id", Value: source_id}, {Key: "account_id", Value: account_id}})
	if err != nil {
		return nil, err
	}

	results := []*Asset{}

	defer func() {
		cursor.Close(ctx)
	}()

	for cursor.Next(ctx) {
		result := &Asset{
			ID: primitive.NilObjectID,
		}

		err := cursor.Decode(&result)
		if err != nil {
			fmt.Printf("error decoding asset result %+v", err)
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

/*******************************************************************************************
 * Identity Operations
 *******************************************************************************************/

// Resolves an identity from a particular account & thread
// if an identity cannot be found, one will be created and saved
// - accepts primitive.ObjectID's of the account and thread
// - returns a pointer to the identity
func (s *Store) ResolveIdentity(account_id, thread_id primitive.ObjectID) (*Identity, error) {
	collection := s.DB.Collection("identities")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := FindFilterIdentityInThread(thread_id, account_id)

	identity := &Identity{
		ID: primitive.NilObjectID,
	}

	_ = collection.FindOne(ctx, filter).Decode(&identity)

	if identity.ID == primitive.NilObjectID {
		identity = NewIdentity()
		identity.Account = account_id
		identity.Thread = thread_id

		err := s.SaveNewSingle(identity, "identities")
		if err != nil {
			return nil, err
		}
	}

	return identity, nil
}

// Count Results
// - accepts a string of the collection name
// - accepts a bson.D of the filter
// - returns an int64 of the count
//
//	Counts the number of documents matching the given filter in the specified collection.
//	useful for pagination since when we query for results we're only receiving a subset of the total results.
func (s *Store) CountResults(col string, filter bson.D) int64 {
	// var result int64 = 0
	result := int64(0)

	count_options := options.Count().SetMaxTime(5 * time.Second)
	collection := s.DB.Collection(col)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, _ = collection.CountDocuments(ctx, filter, count_options)
	return result
}
