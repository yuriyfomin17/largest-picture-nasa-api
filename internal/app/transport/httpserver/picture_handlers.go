package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/common/server"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
	"net/http"
	"strconv"
)

func (h HttpServer) PostCommandHandler(w http.ResponseWriter, r *http.Request) {
	var request PictureCommand

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		server.BadRequest("invalid-command", err, w, r)
		return
	}

	if request.Sol <= 0 {
		server.BadRequest("invalid-command", errors.New("sol must be positive"), w, r)
		return
	}

	h.largestPictureService.PublishCommand(context.Background(), request.Sol)

	server.RespondOK(map[string]interface{}{
		"message": "Command accepted. Largest picture calculation has started.",
	}, w)
}

func (h HttpServer) GetLargestPictureHandler(w http.ResponseWriter, r *http.Request) {
	solStr := mux.Vars(r)["sol"] // Extract path variable
	sol, err := strconv.Atoi(solStr)
	if err != nil {
		server.BadRequest("invalid-command", errors.New("invalid sol"), w, r)
		return
	}
	picture, err := h.largestPictureService.GetPictureBySol(context.Background(), sol)
	if err != nil && errors.Is(err, domain.ErrNotFound) {
		server.NotFound("not-found", domain.ErrNotFound, w, r)
		return
	}

	server.RespondOK(map[string]interface{}{
		"sol":     picture.GetSol(),
		"img_src": picture.GetUrl(),
		"size":    picture.GetSize(),
		"message": "Largest picture fetched successfully",
	}, w)
}
