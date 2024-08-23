package handler

import (
	"encoding/json"
	"errors"
	"github.com/crazybolillo/eryth/internal/model"
	"github.com/crazybolillo/eryth/internal/query"
	"github.com/crazybolillo/eryth/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"strconv"
)

type Endpoint struct {
	Service *service.EndpointService
}

func (e *Endpoint) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", e.create)
	r.Get("/", e.list)
	r.Get("/{sid}", e.get)
	r.Delete("/{sid}", e.delete)
	r.Patch("/{sid}", e.update)

	return r
}

// @Summary Get information from a specific endpoint.
// @Param sid path int true "Requested endpoint's sid"
// @Produce json
// @Success 200 {object} model.Endpoint
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints/{sid} [get]
func (e *Endpoint) get(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sid")
	if sid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(sid, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endpoint, err := e.Service.Read(r.Context(), int32(id))
	if errors.Is(err, pgx.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to read endpoint", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	content, err := json.Marshal(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", slog.String("path", r.URL.Path))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
}

// @Summary List existing endpoints.
// @Param page query int false "Zero based page to fetch" default(0)
// @Param pageSize query int false "Max amount of results to be returned" default(10)
// @Produce json
// @Success 200 {object} model.EndpointPage
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints [get]
func (e *Endpoint) list(w http.ResponseWriter, r *http.Request) {
	page, err := query.GetIntOr(r.URL.Query(), "page", 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pageSize, err := query.GetIntOr(r.URL.Query(), "pageSize", 10)
	if err != nil || page < 0 || pageSize < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := e.Service.Paginate(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("Failed to list endpoints", slog.String("path", r.URL.Path), slog.String("msg", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	content, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to marshall endpoint list", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary Create a new endpoint.
// @Accept json
// @Param payload body model.NewEndpoint true "Endpoint's information"
// @Success 201 {object} model.Endpoint
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints [post]
func (e *Endpoint) create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	payload := model.NewEndpoint{
		MaxContacts: 1,
	}

	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endpoint, err := e.Service.Create(r.Context(), payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to create endpoint", slog.String("reason", err.Error()))
		return
	}

	content, err := json.Marshal(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", slog.String("path", r.URL.Path))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
	w.WriteHeader(http.StatusCreated)
}

// @Summary Delete an endpoint and its associated resources.
// @Param sid path int true "Sid of the endpoint to be deleted"
// @Success 204
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints/{sid} [delete]
func (e *Endpoint) delete(w http.ResponseWriter, r *http.Request) {
	urlSid := chi.URLParam(r, "sid")
	sid, err := strconv.Atoi(urlSid)
	if err != nil || sid <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = e.Service.Delete(r.Context(), int32(sid))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to delete endpoint", slog.String("reason", err.Error()))
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Update the specified endpoint. Omitted or null fields will remain unchanged.
// @Param sid path int true "Sid of the endpoint to be updated"
// @Success 200 {object} model.PatchedEndpoint
// @Failure 400
// @Failure 404
// @Failure 500
// @Tags endpoints
// @Router /endpoints/{sid} [patch]
func (e *Endpoint) update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var payload model.PatchedEndpoint

	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlSid := chi.URLParam(r, "sid")
	sid, err := strconv.Atoi(urlSid)
	if err != nil || sid <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endpoint, err := e.Service.Update(r.Context(), int32(sid), payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to update endpoint", slog.String("reason", err.Error()))
		return
	}

	content, err := json.Marshal(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", slog.String("path", r.URL.Path))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
	w.WriteHeader(http.StatusOK)
}
