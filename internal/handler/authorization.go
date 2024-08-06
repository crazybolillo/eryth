package handler

import (
	"context"
	"encoding/json"
	"github.com/crazybolillo/eryth/internal/bouncer"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type CallBouncer interface {
	Check(ctx context.Context, endpoint, dialed string) bouncer.Response
}

type Authorization struct {
	Bouncer CallBouncer
}

type AuthorizationRequest struct {
	From      string `json:"endpoint"`
	Extension string `json:"extension"`
}

func (e *Authorization) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", e.post)

	return r
}

// @Summary Determine whether the specified action (call) is allowed or not and provide details on how
// to accomplish it.
// @Accept json
// @Produce json
// @Param payload body AuthorizationRequest true "Action to be reviewed"
// @Success 200 {object} bouncer.Response
// @Failure 400
// @Failure 500
// @Tags bouncer
// @Router /bouncer [post]
func (e *Authorization) post(w http.ResponseWriter, r *http.Request) {
	var payload AuthorizationRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := e.Bouncer.Check(r.Context(), payload.From, payload.Extension)
	content, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
	w.WriteHeader(http.StatusOK)
}
