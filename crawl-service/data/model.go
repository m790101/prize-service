package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var client *mongo.Client

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models{
		RestaurantEntry: RestaurantEntry{},
	}

}

type Models struct {
	RestaurantEntry RestaurantEntry
}

type RestaurantEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Rating    float64   `json:"rating"`
	PlaceID   string    `json:"place_id"`
	Area      string    `json:"area"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func (r *RestaurantEntry) EnsureUniqueIndex() error {
	collection := client.Database("restaurants").Collection("restaurants")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "placeid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// Index might already exist, check if it's a "already exists" error
		log.Panicln("err", err)
		if !mongo.IsDuplicateKeyError(err) {
			log.Printf("Error creating index: %v", err)
			return err
		}
	}

	return nil
}

func (r *RestaurantEntry) Insert(entry RestaurantEntry) error {
	collection := client.Database("restaurants").Collection("restaurants")

	_, err := collection.InsertOne(context.TODO(), RestaurantEntry{
		Name:      entry.Name,
		Address:   entry.Address,
		Rating:    entry.Rating,
		PlaceID:   entry.PlaceID,
		Area:      entry.Area,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error inserting into logs:", err)
		return err
	}

	return nil
}

func (r *RestaurantEntry) InsertMany(entrys []RestaurantEntry) error {
	collection := client.Database("restaurants").Collection("restaurants")

	opts := options.InsertMany().SetOrdered(false)
	_, err := collection.InsertMany(context.TODO(), entrys, opts)

	if err != nil {
		// Check if it's just duplicate key errors (which we can ignore)
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Some duplicates skipped (this is normal)")
			return nil
		}
		log.Println("Error inserting restaurants:", err)
		return err
	}

	return nil
}

func (l *RestaurantEntry) All() ([]*RestaurantEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()
	collection := client.Database("restaurants").Collection("restaurants")

	opts := options.Find()

	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)

	if err != nil {
		log.Println("Finding all doc error", err)
	}
	defer cursor.Close(ctx)

	var restaurants []*RestaurantEntry

	for cursor.Next(ctx) {
		var item RestaurantEntry

		err := cursor.Decode(&item)
		if err != nil {
			log.Println("Error decoding restaurant into slice", err)
			return nil, err
		} else {
			restaurants = append(restaurants, &item)
		}
	}
	return restaurants, nil
}

func (r *RestaurantEntry) GetOne(id string) (*RestaurantEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()
	collection := client.Database("restaurants").Collection("restaurants")

	docId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var entry RestaurantEntry

	err = collection.FindOne(ctx, bson.M{"_id": docId}).Decode(&entry)

	if err != nil {
		return nil, err
	}
	return &entry, nil

}

// func (l *LogEntry) DropCollection() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

// 	defer cancel()
// 	collection := client.Database("logs").Collection("logs")

// 	if err := collection.Drop(ctx); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (l *RestaurantEntry) Update() (*mongo.UpdateResult, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

// 	defer cancel()
// 	collection := client.Database("logs").Collection("logs")

// 	docId, err := bson.ObjectIDFromHex(l.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	result, err := collection.UpdateOne(ctx,
// 		bson.M{"_id": docId},
// 		bson.D{
// 			{
// 				"$set", bson.D{
// 					{"name", l.Name},
// 					{"data", l.Data},
// 					{"updated_at", time.Now()},
// 				}},
// 		},
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil

// }
