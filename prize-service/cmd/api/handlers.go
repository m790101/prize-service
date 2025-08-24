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
		Data:    uId,
	}

	app.writeJson(w, http.StatusAccepted, payload)

}

func (app *Config) DrawPrizes(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	var reqestPayload struct {
		Uid string `json:"uId"`
	}

	err := app.readJson(w, r, &reqestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	drawId := fmt.Sprintf("draw:%s", reqestPayload.Uid)

	prize, err := app.Rdb.RPop(ctx, drawId).Result()
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := JsonResponse{
		Status:  "200",
		Message: "",
		Data:    prize,
	}

	app.writeJson(w, http.StatusAccepted, payload)
}
