package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models{
		LogEntry: LogEntry{},
	}
}

type Models struct {
	LogEntry LogEntry
}

// bson is what is used in mongo(binary json)
type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// These are all database function for communicating with our MongoDB

func (l *LogEntry) Insert(entry LogEntry) error {
	collection := client.Database("logs").Collection("log_entries")

	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error inserting log entry: ", err)
		return err
	}

	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("log_entries")

	// Options
	opts := options.Find()
	// Sort the return
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Println("Error finding log entries: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logEntries []*LogEntry

	for cursor.Next(ctx) {
		var logEntry LogEntry
		cursor.Decode(&logEntry)
		if err != nil {
			log.Println("Error decoding log entry: ", err)
			return nil, err
		} else {
			logEntries = append(logEntries, &logEntry)
		}
	}

	return logEntries, nil
}

func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("log_entries")

	docId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var logEntry LogEntry
	err = collection.FindOne(ctx, bson.M{"_id": docId}).Decode(&logEntry)
	if err != nil {
		log.Println("Error finding log entry: ", err)
		return nil, err
	}

	return &logEntry, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("log_entries")

	if err := collection.Drop(ctx); err != nil {
		log.Println("Error dropping log entry collection: ", err)
		return err
	}

	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("log_entries")

	docId, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		return nil, err
	}

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": docId},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: l.Name}, 
				{Key: "data", Value: l.Data}, 
				{Key: "updated_at", Value: time.Now()},
			}},
		},
	)
	if err != nil {
		log.Println("Error updating log entry: ", err)
		return nil, err
	}

	return result, nil
}
