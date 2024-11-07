package handler

import (
	"encoding/json"
	"github.com/crazybolillo/eryth/internal/query"
	"github.com/crazybolillo/eryth/internal/service"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type Cdr struct {
	Service *service.Cdr
}

func (c *Cdr) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", c.list)

	return r
}

// @Summary List call detail records.
// @Param page query int false "Zero based page to fetch" default(0)
// @Param pageSize query int false "Max amount of results to be returned" default(20)
// @Produce json
// @Success 200 {object} model.CallRecordPage
// @Tags cdr
// @Router /cdr [get]
func (c *Cdr) list(w http.ResponseWriter, r *http.Request) {
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

	res, err := c.Service.Paginate(r.Context(), page, pageSize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to list cdr", "path", r.URL.Path, "reason", err)
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
