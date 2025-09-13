package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"prize-service/data"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	webPort = "80"
)

type Config struct {
	Rdb    *redis.Client
	Models data.Models
}

func main() {
	log.Println("prize-service init")

	redisClient, err := connectToRedis()
	if err != nil {
		log.Panic(err)
	}

	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	app := Config{
		Rdb:    redisClient,
		Models: data.New(mongoClient),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}

}

func connectToRedis() (*redis.Client, error) {

	var ctx = context.Background()

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisUsername := os.Getenv("REDIS_USERNAME")

	clientOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Username: redisUsername,
		Password: redisPassword,
		DB:       0,
	}

	rdb := redis.NewClient(clientOptions)
	_, err := rdb.Ping(ctx).Result()

	// connect
	if err != nil {
		log.Panic("Error connection", err)
		return nil, err
	}

	log.Println("connected to redis")

	return rdb, nil
}

func connectToMongo() (*mongo.Client, error) {
	mongoUrl := os.Getenv("MONGO_URL")

	if mongoUrl == "" {
		mongoUrl = "mongodb://mongo:27017"
	}

	clientOptions := options.Client().ApplyURI(mongoUrl)
	// clientOptions.SetAuth(options.Credential{
	// 	Username: "admin",
	// 	Password: "password",
	// })

	// connect
	c, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Panic("Error connection", err)
		return nil, err
	}

	log.Println("connected to mongo")

	return c, nil
}
