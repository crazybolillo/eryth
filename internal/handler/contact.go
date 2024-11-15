package handler

import (
	"encoding/json"
	"github.com/crazybolillo/eryth/internal/query"
	"github.com/crazybolillo/eryth/internal/service"
	"github.com/crazybolillo/eryth/pkg/model"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strings"
)

type Contact struct {
	Service *service.Contact
}

func (p *Contact) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", p.list)

	return r
}

// @Summary List all contacts in the system.
// @Param page query int false "Zero based page to fetch" default(0)
// @Param pageSize query int false "Max amount of results to be returned" default(20)
// @Param name query string false "Filter by name. Asterisks can be used to represent any character(s)"
// @Param phone query string false "Filter by phone number. Asterisks can be used to represent any character(s)"
// @Param op query string false "Whether to join filter values with AND/OR logic." default(and)
// @Produce json
// @Success 200 {object} model.ContactPage
// @Tags contacts
// @Router /contacts [get]
func (p *Contact) list(w http.ResponseWriter, r *http.Request) {
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

	res, err := p.Service.Paginate(
		r.Context(),
		model.ContactPageFilter{
			Name:     strings.ReplaceAll(r.URL.Query().Get("name"), "*", "%"),
			Phone:    strings.ReplaceAll(r.URL.Query().Get("phone"), "*", "%"),
			Operator: strings.ToLower(r.URL.Query().Get("op")),
		},
		page,
		pageSize,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to list contacts", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	content, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
}
