package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

const (
	webPort = "80"
)

type Config struct {
	Rdb *redis.Client
}

func main() {
	log.Println("prize-service init")

	redisClient, err := connectToRedis()
	if err != nil {
		log.Panic(err)
	}

	app := Config{
		Rdb: redisClient,
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
