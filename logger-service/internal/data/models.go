package data

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoTimeout = 15 * time.Second

var (
	mongoClient *mongo.Client
	dbName      string
	collName    string
)

// EntryInserter persists log entries.
type EntryInserter interface {
	Insert(ctx context.Context, entry LogEntry) error
}

func New(client *mongo.Client, db, collection string) Models {
	mongoClient = client
	dbName = db
	collName = collection

	return Models{
		LogEntry: LogEntry{},
	}
}

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func (l *LogEntry) Insert(ctx context.Context, entry LogEntry) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), mongoTimeout)
		defer cancel()
	}

	collection := mongoClient.Database(dbName).Collection(collName)

	_, err := collection.InsertOne(ctx, LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		slog.Error("insert log entry", "err", err)
		return err
	}

	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	collection := mongoClient.Database(dbName).Collection(collName)

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		slog.Error("find log entries", "err", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*LogEntry

	for cursor.Next(ctx) {
		var item LogEntry
		if err := cursor.Decode(&item); err != nil {
			slog.Error("decode log entry", "err", err)
			return nil, err
		}
		logs = append(logs, &item)
	}

	return logs, nil
}

func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	collection := mongoClient.Database(dbName).Collection(collName)

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var entry LogEntry
	if err := collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	collection := mongoClient.Database(dbName).Collection(collName)
	return collection.Drop(ctx)
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	collection := mongoClient.Database(dbName).Collection(collName)

	docID, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		return nil, err
	}

	return collection.UpdateOne(
		ctx,
		bson.M{"_id": docID},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: l.Name},
				{Key: "data", Value: l.Data},
				{Key: "updated_at", Value: time.Now()},
			}},
		},
	)
}
