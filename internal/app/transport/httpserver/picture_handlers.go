package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/common/server"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
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

	err := h.largestPictureService.PublishCommand(request.Sol)

	if err != nil {
		server.BadRequest("could-not-publish-command", err, w, r)
		return
	}
	server.RespondOK(server.SuccessResponse{
		Message: "Command accepted. Largest picture calculation has started.",
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
	server.RespondOK(server.SuccessResponse{
		Sol:     picture.GetSol(),
		ImgSrc:  picture.GetUrl(),
		Size:    picture.GetSize(),
		Message: "Command accepted. Largest picture calculation has started.",
	}, w)
}
