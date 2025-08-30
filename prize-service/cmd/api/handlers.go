package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

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
