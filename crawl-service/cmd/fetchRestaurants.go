package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Restaurant struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Rating  float64 `json:"rating"`
	PlaceID string  `json:"place_id"`
	Area    string  `json:"area"`
}

type SearchRequest struct {
	IncludedTypes       []string            `json:"includedTypes"`
	MaxResultCount      int                 `json:"maxResultCount"`
	LocationRestriction LocationRestriction `json:"locationRestriction"`
	LanguageCode        string              `json:"languageCode"`
	RankPreference      string              `json:"rankPreference,omitempty"`
}

type LocationRestriction struct {
	Circle Circle `json:"circle"`
}

type Circle struct {
	Center Center  `json:"center"`
	Radius float64 `json:"radius"`
}

type Center struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type SearchResponse struct {
	Places []Place `json:"places"`
}

type Place struct {
	ID               string      `json:"id"`
	DisplayName      DisplayName `json:"displayName"`
	FormattedAddress string      `json:"formattedAddress"`
	Rating           float64     `json:"rating"`
}

type DisplayName struct {
	Text         string `json:"text"`
	LanguageCode string `json:"languageCode"`
}

func fetchRestaurantsAtPoint(lat, lng float64, pointName string, rankBy string, searchNum int) ([]Restaurant, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var googleKey = os.Getenv("GOOGLE_KEY")

	log.Println("googleKey", googleKey)
	// Add random offset to coordinates for variety (±300m)
	offsetLat := lat + (rand.Float64()-0.5)*0.006
	offsetLng := lng + (rand.Float64()-0.5)*0.006

	// Vary radius for different searches
	radiuses := []float64{500, 900, 1300}
	radius := radiuses[searchNum%len(radiuses)]

	requestBody := SearchRequest{
		IncludedTypes: []string{
			"restaurant",
			"meal_takeaway",
			"bakery",
			"cafe",
			"meal_delivery",
		},
		MaxResultCount: 20,
		LocationRestriction: LocationRestriction{
			Circle: Circle{
				Center: Center{
					Latitude:  offsetLat,
					Longitude: offsetLng,
				},
				Radius: radius,
			},
		},
		LanguageCode:   "zh-TW",
		RankPreference: rankBy,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", googleKey)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.formattedAddress,places.rating")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	restaurants := make([]Restaurant, 0, len(searchResp.Places))
	for _, place := range searchResp.Places {
		restaurant := Restaurant{
			Name:    place.DisplayName.Text,
			Address: place.FormattedAddress,
			Rating:  place.Rating,
			PlaceID: place.ID,
			Area:    pointName,
		}
		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}
