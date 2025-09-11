package main

import (
	"context"
	"crawl-service/data"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	mongoUrl = "mongodb://localhost:27017"
	API_URL  = "https://places.googleapis.com/v1/places:searchNearby"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

var searchPoints = []struct {
	Name string
	Lat  float64
	Lng  float64
}{
	{"daanPark", 25.0263, 121.5349},
	{"Zhongxiao/Fuxing", 25.0417, 121.5436},
	{"Xinyi/Jianguo", 25.0337, 121.5364},
	{"Heping/Xinsheng", 25.0266, 121.5279},
	{"Ren'ai/Dunhua", 25.0378, 121.5485},
	{"Zhongshan/Jianguo", 25.0445, 121.5364},
	{"Xinyi/Keelung", 25.0473, 121.5617},
	{"Heping/Jiangnan", 25.0200, 121.5400},
}

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoUrl)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect
	c, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Panic("Error connection", err)
		return nil, err
	}

	log.Println("connected to mongo")

	return c, nil
}

func main() {

	mongoClient, err := connectToMongo()
	if err != nil {
		log.Println("error connect to mongo db")
		return
	}

	client = mongoClient

	// create context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	err = app.Models.RestaurantEntry.EnsureUniqueIndex()
	if err != nil {
		log.Printf("Error creating index: %v", err)
	}

	fmt.Printf("Searching %d different areas in Daan District with coordinate variation...\n\n", len(searchPoints))

	allRestaurants := make(map[string]Restaurant)

	rankMethods := []string{"POPULARITY", "DISTANCE"}

	for i, point := range searchPoints {
		rankBy := rankMethods[i%len(rankMethods)]
		fmt.Printf("Searching %s (%.4f, %.4f) - Ranking by %s, Search #%d...\n",
			point.Name, point.Lat, point.Lng, rankBy, i+1)

		restaurants, err := fetchRestaurantsAtPoint(point.Lat, point.Lng, point.Name, rankBy, i)
		if err != nil {
			log.Printf("Error fetching from %s: %v", point.Name, err)
			continue
		}

		newCount := 0
		for _, r := range restaurants {
			if _, exists := allRestaurants[r.PlaceID]; !exists {
				allRestaurants[r.PlaceID] = r
				newCount++
			}
		}

		fmt.Printf("Found %d restaurants (%d new, %d total unique)\n", len(restaurants), newCount, len(allRestaurants))

		if i < len(searchPoints)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	uniqueRestaurants := make([]Restaurant, 0, len(allRestaurants))
	for _, r := range allRestaurants {
		uniqueRestaurants = append(uniqueRestaurants, r)
	}

	fmt.Printf("\n=== FINAL RESULTS ===\n")
	fmt.Printf("Total unique restaurants found: %d\n\n", len(uniqueRestaurants))

	// count := 200
	// if len(uniqueRestaurants) < 30 {
	// 	count = len(uniqueRestaurants)
	// }

	payloads := []data.RestaurantEntry{}

	for i := 0; i < len(uniqueRestaurants); i++ {
		r := uniqueRestaurants[i]

		restaurantPayload := data.RestaurantEntry{
			Name:    r.Name,
			Address: r.Address,
			Rating:  r.Rating,
			PlaceID: r.PlaceID,
			Area:    r.Area,
		}

		payloads = append(payloads, restaurantPayload)

		fmt.Printf("%d. %s\n", i+1, r.Name)
		if r.Rating > 0 {
			fmt.Printf("   Rating: %.1f/5\n", r.Rating)
		}
		fmt.Printf("   Address: %s\n", r.Address)
		fmt.Printf("   Place ID: %s\n\n", r.PlaceID)
	}

	log.Println("start adding data to db")

	app.Models.RestaurantEntry.InsertMany(payloads)

	log.Println("finished adding data to db")
}
