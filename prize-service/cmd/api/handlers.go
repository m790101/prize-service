package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"prize-service/data"
	"time"

	"github.com/google/uuid"
)

type Prizes struct {
	ID   int
	Name string
}

func (app *Config) NewPrizes(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	var reqestPayload struct {
		Names []string `json:"names"`
	}

	err := app.readJson(w, r, &reqestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	names := reqestPayload.Names

	// make Names to a random slice
	rand.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })

	// add to list with randomize names & drawId
	uId := uuid.New()
	drawId := fmt.Sprintf("draw:%s", uId)
	err = app.Rdb.LPush(ctx, drawId, names).Err()
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := JsonResponse{
		Status:  "200",
		Message: "",
		Data: struct {
			ID uuid.UUID `json:"id"`
		}{
			ID: uId,
		},
	}

	app.writeJson(w, http.StatusOK, payload)

}

func (app *Config) DrawPrizes(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	var reqestPayload struct {
		UId string `json:"uId"`
	}

	err := app.readJson(w, r, &reqestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	drawId := fmt.Sprintf("draw:%s", reqestPayload.UId)

	prize, err := app.Rdb.RPop(ctx, drawId).Result()
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := JsonResponse{
		Status:  "200",
		Message: "",
		Data: struct {
			Name string `json:"name"`
		}{
			Name: prize,
		},
	}

	app.writeJson(w, http.StatusOK, payload)
}

func (app *Config) UpdatePrized(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var reqestPayload struct {
		UId   string   `json:"uId"`
		Names []string `json:"names"`
	}

	err := app.readJson(w, r, &reqestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	names := reqestPayload.Names

	// make Names to a random slice
	rand.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })

	// add to list with randomize names & drawId
	drawId := fmt.Sprintf("draw:%s", reqestPayload.UId)
	app.Rdb.Del(ctx, drawId)

	err = app.Rdb.LPush(ctx, drawId, names).Err()
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := JsonResponse{
		Status:  "200",
		Message: "",
		Data:    struct{}{},
	}

	app.writeJson(w, http.StatusOK, payload)

}

func (app *Config) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	res := JsonResponse{
		Status:  "9999",
		Message: "route not found",
		Data:    struct{}{},
	}
	app.writeJson(w, http.StatusNotFound, res)
}

func (app *Config) DrawRestaurants(w http.ResponseWriter, r *http.Request) {
	log.Println("draw restaurant")
	ctx := context.Background()
	redisKey := "restaurants"

	// Check if there is data in Redis
	exists, err := app.Rdb.Exists(ctx, redisKey).Result()
	if err != nil {
		app.errorJson(w, err)
		return
	}

	log.Println("exists", exists)

	var restaurants []*data.RestaurantEntry

	if exists == 0 {
		// No data in Redis - get restaurants from database
		restaurantData, err := app.Models.RestaurantEntry.All()
		if err != nil {
			app.errorJson(w, err)
			return
		}

		restaurants = restaurantData

		if len(restaurants) > 0 {
			restaurantBytes, err := json.Marshal(restaurants)
			if err != nil {
				app.errorJson(w, err)
				return
			}

			// Put restaurants JSON in Redis for future use
			err = app.Rdb.Set(ctx, redisKey, restaurantBytes, time.Hour).Err()
			if err != nil {
				app.errorJson(w, err)
				return
			}
		}
	} else {
		// Get restaurants from Redis
		restaurantData, err := app.Rdb.Get(ctx, redisKey).Result()
		if err != nil {
			app.errorJson(w, err)
			return
		}
		// Unmarshal JSON data back to restaurant objects
		err = json.Unmarshal([]byte(restaurantData), &restaurants)
		if err != nil {
			app.errorJson(w, err)
			return
		}
	}

	if len(restaurants) == 0 {
		app.errorJson(w, fmt.Errorf("no restaurants available"))
		return
	}

	// Shuffle the restaurants
	rand.Shuffle(len(restaurants), func(i, j int) {
		restaurants[i], restaurants[j] = restaurants[j], restaurants[i]
	})

	// Select up to 3 restaurants
	maxSelect := 3
	if len(restaurants) < maxSelect {
		maxSelect = len(restaurants)
	}
	selectedRestaurants := restaurants[:maxSelect]

	type RestaurantRes struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		Address string  `json:"address"`
		Rating  float64 `json:"rating"`
		PlaceID string  `json:"placeId"`
		Area    string  `json:"area"`
	}

	var responseRestaurants []RestaurantRes
	for _, restaurant := range selectedRestaurants {
		responseRestaurants = append(responseRestaurants, RestaurantRes{
			ID:      restaurant.ID.Hex(),
			Name:    restaurant.Name,
			Address: restaurant.Address,
			Rating:  restaurant.Rating,
			PlaceID: restaurant.PlaceID,
			Area:    restaurant.Area,
		})
	}

	// Return response
	payload := JsonResponse{
		Status:  "200",
		Message: "",
		Data: struct {
			Restaurants []RestaurantRes `json:"restaurants"`
		}{
			Restaurants: responseRestaurants,
		},
	}

	app.writeJson(w, http.StatusOK, payload)
}
