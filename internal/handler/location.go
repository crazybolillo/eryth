package handler

import (
	"encoding/json"
	"github.com/crazybolillo/eryth/internal/query"
	"github.com/crazybolillo/eryth/internal/service"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type Location struct {
	Service *service.Location
}

func (l *Location) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", l.list)

	return r
}

// @Summary List endpoint locations. This is the same as PJSIP contacts.
// @Param page query int false "Zero based page to fetch" default(0)
// @Param pageSize query int false "Max amount of results to be returned" default(20)
// @Produce json
// @Success 200 {object} model.LocationPage
// @Tags location
// @Router /locations [get]
func (l *Location) list(w http.ResponseWriter, r *http.Request) {
	page, err := query.GetIntOr(r.URL.Query(), "page", 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pageSize, err := query.GetIntOr(r.URL.Query(), "pageSize", 25)
	if err != nil || page < 0 || pageSize < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := l.Service.Paginate(r.Context(), page, pageSize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to list locations", "path", r.URL.Path, "reason", err)
		return
	}

	content, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", "path", r.URL.Path, "reason", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", "path", r.URL.Path, "reason", err)
	}
}
